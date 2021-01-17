package pkg

import (
	"Unison-Docker-Controller/api/types/config_types"
	"Unison-Docker-Controller/api/types/container_types"
	"Unison-Docker-Controller/internal/container_internal"
	"Unison-Docker-Controller/internal/local_sys"
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerController struct {
	Config      config_types.Config
	SysBaseInfo *local_sys.SystemBaseInfo
	SysResource *local_sys.SystemResource

	CCB map[string]*container_internal.ContainerControlBlock `json:"container control block"`

	cli *client.Client
}

func NewDockerController(cfg config_types.Config) (*DockerController, error) {
	// TODO 如何处理现有的其余 docker container

	sysBaseInfo, errSBI := local_sys.NewSystemBaseInfo(cfg)
	if errSBI != nil {
		return nil, errSBI
	}

	sysDynamicInfo, errSDI := local_sys.NewSystemResource(cfg, sysBaseInfo.LogicalCores)
	if errSDI != nil {
		return nil, errSDI
	}

	dockerClient, errCli := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if errCli != nil {
		return nil, errCli
	}

	c := &DockerController{
		Config:      cfg,
		SysBaseInfo: sysBaseInfo,
		SysResource: sysDynamicInfo,
		CCB:         make(map[string]*container_internal.ContainerControlBlock),
		cli:         dockerClient,
	}

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

	resp, err := ctr.cli.ContainerCreate(context.Background(),
		&container.Config{
			Image:        cfg.ImageName,
			ExposedPorts: exports,
			Tty:          true,
			StopTimeout:  &ctr.Config.ContainerStopTimeout,
		}, nil, nil, nil, cfg.ContainerName)

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

//statusCh, errCh := ctr.cli.ContainerWait(context.Background(), resp.ID, container.WaitConditionNotRunning)
//select {
//case err := <-errCh:
//if err != nil {
//panic(err)
//}
//case info := <-statusCh:
//print(info)
//}
//
//out, err := ctr.cli.ContainerLogs(context.Background(), resp.ID, types.ContainerLogsOptions{ShowStdout: true})
//if err != nil {
//panic(err)
//}
//
//_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
