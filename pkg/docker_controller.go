package pkg

import (
	"Unison-Docker-Controller/internal/config"
	"Unison-Docker-Controller/internal/local_sys"
)

type DockerController struct {
	Config         config.Config
	SysBaseInfo    *local_sys.SystemBaseInfo
	SysDynamicInfo *local_sys.SystemDynamicInfo
}

func NewDockerController(cfg config.Config) (*DockerController, error) {
	c := &DockerController{
		Config:         cfg,
		SysBaseInfo:    nil,
		SysDynamicInfo: nil,
	}

	sysBaseInfo, errSBI := local_sys.NewSystemBaseInfo(cfg)
	if errSBI != nil {
		return nil, errSBI
	}

	sysDynamicInfo, errSDI := local_sys.NewSystemDynamicInfo(cfg)
	if errSDI != nil {
		return nil, errSDI
	}

	c.SysBaseInfo = sysBaseInfo
	c.SysDynamicInfo = sysDynamicInfo

	return c, nil
}

func (ctr *DockerController) UpdateInfo() error {
	err := ctr.SysDynamicInfo.UpdateStatus(ctr.Config)
	return err
}
