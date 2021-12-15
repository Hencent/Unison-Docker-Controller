package controller

import (
	"context"
	types2 "github.com/PenguinCats/Unison-Docker-Controller/api/types"
)

func (ctr *DockerController) ContainerStop(ExtContainerID string) error {
	ccb, err := ctr.getCCB(ExtContainerID)
	if err != nil {
		return err
	}

	err = ctr.cli.ContainerStop(context.Background(), ccb.ContainerID, &ctr.containerStopTimeout)
	if err != nil {
		return types2.ErrInternalError
	}

	ctr.releaseRunningResourceForContainer(ccb)

	return nil
}
