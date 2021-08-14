package controller

import (
	"context"
	"fmt"
)

func (ctr *DockerController) ContainerStop(containerID string) error {
	if !ctr.ContainerIsExist(containerID) {
		return fmt.Errorf("container [%s] does not exist", containerID)
	}

	ccb, err := ctr.getCCB(containerID)
	if err != nil {
		return err
	}

	err = ctr.cli.ContainerStop(context.Background(), containerID, nil)
	if err != nil {
		return err
	}

	ctr.releaseRunningResourceForContainer(ccb)

	return nil
}
