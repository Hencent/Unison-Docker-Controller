package pkg

import (
	"Unison-Docker-Controller/api/types/config"
	container2 "Unison-Docker-Controller/api/types/container"
	"Unison-Docker-Controller/internal/local_sys"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerController struct {
	Config         config.Config
	SysBaseInfo    *local_sys.SystemBaseInfo
	SysDynamicInfo *local_sys.SystemDynamicInfo

	cli *client.Client
}

func NewDockerController(cfg config.Config) (*DockerController, error) {
	c := &DockerController{
		Config: cfg,
	}

	sysBaseInfo, errSBI := local_sys.NewSystemBaseInfo(cfg)
	if errSBI != nil {
		return nil, errSBI
	}

	sysDynamicInfo, errSDI := local_sys.NewSystemDynamicInfo(cfg)
	if errSDI != nil {
		return nil, errSDI
	}

	dockerClient, errCli := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if errCli != nil {
		return nil, errCli
	}

	c.SysBaseInfo = sysBaseInfo
	c.SysDynamicInfo = sysDynamicInfo
	c.cli = dockerClient

	return c, nil
}

func (ctr *DockerController) UpdateInfo() error {
	err := ctr.SysDynamicInfo.UpdateStatus(ctr.Config)
	return err
}

func (ctr *DockerController) ContainerCreat(cfg container2.ContainerConfig) (string, error) {
	exports, errExports := generateExportsForContainer(cfg.ExposedTCPPorts, cfg.ExposedUDPPorts)
	if errExports != nil {
		return "", errExports
	}

	resp, err := ctr.cli.ContainerCreate(context.Background(),
		&container.Config{
			Image:        cfg.ImageName,
			ExposedPorts: exports,
			Tty:          true,
			AttachStdin:  true,
		}, nil, nil, nil, cfg.ContainerName)

	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// TODO 资源限制在启动前使用 docker update 来限制
//&container.Config{
//Cmd:   []string{"echo", "hello world"},
//Image: cfg.ImageName,
//ExposedPorts: exports,
//Tty:   true,
//AttachStdin: true,
//},
//&container.HostConfig{
//Resources: container.Resources{
//Memory: int64(cfg.RamAmount),
//},
//},

//if err := ctr.cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
//panic(err)
//}
//
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
