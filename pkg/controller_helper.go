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
	errUsage := ctr.containerUpdateAllResourceUsage()
	if errUsage != nil {
		return errUsage
	}

	errResource := ctr.updateDynamicResource()
	if errResource != nil {
		return errResource
	}

	return nil
}
