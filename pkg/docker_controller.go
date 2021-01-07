package pkg

import "Unison-Docker-Controller/internal/local_sys"

type DockerController struct {
	SysBaseInfo *local_sys.SystemBaseInfo
}

func NewDockerController() (*DockerController, error) {
	sysInfo, err := local_sys.NewLocalSystemInfo()
	if err != nil {
		return nil, err
	}

	c := &DockerController{
		SysBaseInfo: sysInfo,
	}

	return c, nil
}
