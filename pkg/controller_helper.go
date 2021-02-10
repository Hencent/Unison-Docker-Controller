package pkg

import (
	"time"
)

func (ctr *DockerController) beginPeriodicTask() {
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(ctr.Config.PeriodicSystemStatsUpdateInterval))

		for range ticker.C {
			_ = ctr.systemStatsUpdate()
		}
	}()
}

// 周期性调度执行
func (ctr *DockerController) systemStatsUpdate() error {
	errContainerUsage := ctr.containerUpdateAllResourceUsage()
	if errContainerUsage != nil {
		return errContainerUsage
	}

	errVolumeUsage := ctr.volumeUpdateAllResourceUsage()
	if errVolumeUsage != nil {
		return errVolumeUsage
	}

	errResource := ctr.updateDynamicResource()
	if errResource != nil {
		return errResource
	}

	return nil
}
