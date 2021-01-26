package pkg

import "time"

func (ctr *DockerController) beginPeriodicTask(periodicTaskInterval int) {
	ticker := time.NewTicker(time.Second * time.Duration(periodicTaskInterval))

	go func() {
		select {
		case <-ticker.C:
			go ctr.SystemStatsUpdate()
		}
	}()
}
