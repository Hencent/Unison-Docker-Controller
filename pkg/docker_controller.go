package pkg

import (
	"Unison-Docker-Controller/api/types/config_types"
	"Unison-Docker-Controller/api/types/container_types"
	"Unison-Docker-Controller/api/types/local_sys_types"
	"Unison-Docker-Controller/api/types/volume_types"
	"Unison-Docker-Controller/internal/container_internal"
	"Unison-Docker-Controller/internal/local_sys"
	"Unison-Docker-Controller/internal/volume_internal"
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"path"
)

type DockerController struct {
	Config config_types.Config

	SysBaseInfo *local_sys_types.SystemBaseInfo
	SysResource *local_sys.SystemResourceController

	CCB map[string]*container_internal.ContainerControlBlock `json:"container control block"`

	VCB map[string]*volume_internal.VolumeControlBlock `json:"volume control block"`

	cli *client.Client
}

func NewDockerController(cfg config_types.Config) (*DockerController, error) {
	dockerClient, errCli := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if errCli != nil {
		return nil, errCli
	}

	containerStopAndRemoveAllForInit(dockerClient, cfg.StopExistingContainersOnStart, cfg.RemoveExistingContainersOnStart)

	dockerRootDir, errDir := getDockerRootDir(dockerClient)
	if errDir != nil {
		return nil, errDir
	}

	sysBaseInfo, errSBI := local_sys_types.NewSystemBaseInfo(dockerRootDir)
	if errSBI != nil {
		return nil, errSBI
	}

	sysResource, errSDI := local_sys.NewSystemResourceController(cfg, sysBaseInfo.LogicalCores, dockerRootDir)
	if errSDI != nil {
		return nil, errSDI
	}

	c := &DockerController{
		Config:      cfg,
		SysBaseInfo: sysBaseInfo,
		SysResource: sysResource,
		CCB:         make(map[string]*container_internal.ContainerControlBlock),
		VCB:         make(map[string]*volume_internal.VolumeControlBlock),
		cli:         dockerClient,
	}

	c.beginPeriodicTask()

	return c, nil
}

func (ctr *DockerController) ContainerIsExist(containerID string) bool {
	_, ok := ctr.CCB[containerID]
	return ok
}

func (ctr *DockerController) ContainerGetStatus(containerID string) container_internal.ContainerStatus {
	if !ctr.ContainerIsExist(containerID) {
		return container_internal.Error
	}

	inspectRes, err := ctr.cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return container_internal.Error
	}

	status := container_internal.Error
	switch inspectRes.State.Status {
	case "created":
		status = container_internal.Created
	case "running":
		status = container_internal.Running
	case "paused":
		status = container_internal.Paused
	case "restarting":
		status = container_internal.Running
	case "removing":
		status = container_internal.Running
	case "exited":
		status = container_internal.Running
	case "dead":
		status = container_internal.Running
	default:
		status = container_internal.Error
	}

	return status
}

func (ctr *DockerController) ContainerCreat(cfg container_types.ContainerConfig) (string, error) {
	exports, errExports := generateExportsForContainer(cfg.ExposedTCPPorts, cfg.ExposedUDPPorts)
	if errExports != nil {
		return "", errExports
	}

	mountInfo := make([]mount.Mount, len(cfg.Volumes))
	for k, v := range cfg.Volumes {
		mountInfo[k] = mount.Mount{
			Type:   "volume",
			Source: v,
			Target: path.Join("/volume", v),
		}
	}

	resp, err := ctr.cli.ContainerCreate(context.Background(),
		&container.Config{
			Image:        cfg.ImageName,
			ExposedPorts: exports,
			Tty:          true,
			StopTimeout:  &ctr.Config.ContainerStopTimeout,
		}, &container.HostConfig{
			Mounts:     mountInfo,
			StorageOpt: map[string]string{},
		}, nil, nil, cfg.ContainerName)

	if err != nil {
		return "", err
	}

	ctr.CCB[resp.ID] = &container_internal.ContainerControlBlock{
		Status: container_internal.Created,
		Config: cfg,
	}

	ctr.containerUpdateStatus(resp.ID)
	return resp.ID, nil
}

func (ctr *DockerController) ContainerStart(containerID string) error {
	if !ctr.ContainerIsExist(containerID) {
		return errors.New("container not exits")
	}

	errUpdateDynamicResource := ctr.updateDynamicResource()
	if errUpdateDynamicResource != nil {
		return errUpdateDynamicResource
	}

	errResource := ctr.containerRequestResource(containerID)
	if errResource != nil {
		return errResource
	}

	err := ctr.cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	if err != nil {
		ctr.CCB[containerID].Status = ctr.ContainerGetStatus(containerID)
		return err
	}

	ctr.containerUpdateStatus(containerID)
	return nil
}

func (ctr *DockerController) ContainerStop(containerID string) error {
	if !ctr.ContainerIsExist(containerID) {
		return errors.New("container not exits")
	}

	err := ctr.cli.ContainerStop(context.Background(), containerID, nil)
	if err != nil {
		ctr.CCB[containerID].Status = ctr.ContainerGetStatus(containerID)
		return err
	}

	ctr.containerReleaseResource(containerID)

	ctr.containerUpdateStatus(containerID)
	return nil
}

func (ctr *DockerController) ContainerRemove(containerID string) error {
	if !ctr.ContainerIsExist(containerID) {
		return errors.New("container not exits")
	}

	err := ctr.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
		//RemoveVolumes: true,
		RemoveVolumes: false,
		Force:         true,
	})
	if err != nil {
		return err
	}

	delete(ctr.CCB, containerID)

	return nil
}

func (ctr *DockerController) ContainerResourceUsage(containerID string) (*container_types.ContainerResourceUsage, error) {
	if !ctr.ContainerIsExist(containerID) {
		return nil, errors.New("container not exits")
	}

	return &ctr.CCB[containerID].ResourceUsage, nil
}

func (ctr *DockerController) ContainerAllResourceUsage() map[string]*container_types.ContainerResourceUsage {
	usage := make(map[string]*container_types.ContainerResourceUsage)
	for k, v := range ctr.CCB {
		usage[k] = &v.ResourceUsage
	}
	return usage
}

func (ctr *DockerController) VolumeIsExist(volumeName string) bool {
	_, ok := ctr.VCB[volumeName]
	return ok
}

func (ctr *DockerController) VolumeCreate(volumeName string) error {
	_, err := ctr.cli.VolumeCreate(context.Background(), volume.VolumeCreateBody{
		Name: volumeName,
	})
	if err != nil {
		return err
	}

	ctr.VCB[volumeName] = &volume_internal.VolumeControlBlock{}

	return nil
}

func (ctr *DockerController) VolumeRemove(volumeName string, force bool) error {
	refCount, errRef := ctr.VolumeRefCount(volumeName)
	if errRef != nil {
		return errRef
	}
	if refCount != 0 {
		return errors.New("volume is still used by container(s)")
	}

	err := ctr.cli.VolumeRemove(context.Background(), volumeName, force)
	if err != nil {
		return err
	}

	delete(ctr.VCB, volumeName)

	return nil
}

func (ctr *DockerController) VolumeRefCount(volumeName string) (int64, error) {
	if !ctr.VolumeIsExist(volumeName) {
		return 0, errors.New("volume not exists")
	}

	return ctr.VCB[volumeName].ResourceUsage.RefCount, nil
}

func (ctr *DockerController) VolumeResourceUsage(volumeName string) (*volume_types.VolumeResourceUsage, error) {
	if !ctr.VolumeIsExist(volumeName) {
		return nil, errors.New("volume not exists")
	}

	return &ctr.VCB[volumeName].ResourceUsage, nil
}

func (ctr *DockerController) VolumeAllResourceUsage() map[string]*volume_types.VolumeResourceUsage {
	usage := make(map[string]*volume_types.VolumeResourceUsage)
	for k, v := range ctr.VCB {
		usage[k] = &v.ResourceUsage
	}
	return usage
}

func (ctr *DockerController) SystemBaseInfo() *local_sys_types.SystemBaseInfo {
	return ctr.SysBaseInfo
}

func (ctr *DockerController) SystemResource() (*local_sys_types.SystemResourceAvailable, error) {
	err := ctr.updateDynamicResource()
	if err != nil {
		return nil, err
	}

	return ctr.SysResource.GetResourceAvailable(), nil
}
