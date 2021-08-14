package controller

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
)

func (ctr *DockerController) ContainerRemove(containerID string) error {
	if !ctr.ContainerIsExist(containerID) {
		return fmt.Errorf("container [%s] does not exist", containerID)
	}

	err := ctr.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return err
	}

	ctr.containerCtrlBlkMutex.Lock()
	delete(ctr.containerCtrlBlk, containerID)
	ctr.containerCtrlBlkMutex.Unlock()

	return nil
}
