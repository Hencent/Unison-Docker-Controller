package controller

import (
	"context"
	types2 "github.com/PenguinCats/Unison-Docker-Controller/api/types"
	container_controller "github.com/PenguinCats/Unison-Docker-Controller/pkg/controller/internal/container-controller"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
)

func getCoreListString(coreList []string) string {
	coreListString := coreList[0]
	for i := 1; i < len(coreList); i += 1 {
		coreListString += "," + coreList[i]
	}
	return coreListString
}

func (ctr *DockerController) allocateRunningResourceForContainer(ccb *container_controller.ContainerControlBlock) (err error) {
	coreList, err := ctr.resourceCtrl.RunningResourceRequest(ccb.CoreRequest, ccb.MemoryRequest)
	if err != nil {
		return err
	}
	coreListString := getCoreListString(coreList)
	defer func() {
		if err != nil {
			ctr.resourceCtrl.RunningResourceRelease(coreList, ccb.MemoryRequest)
		}
	}()

	_, err = ctr.cli.ContainerUpdate(context.Background(), ccb.ContainerID, container.UpdateConfig{
		Resources: container.Resources{
			Memory:     ccb.MemoryRequest,
			CpusetCpus: coreListString,
		},
		RestartPolicy: container.RestartPolicy{},
	})
	if err != nil {
		logrus.Warning(err.Error())
		return types2.ErrInternalError
	}

	ccb.UpdateRunningResourceAllocated(coreList)

	return nil
}

func (ctr *DockerController) releaseRunningResourceForContainer(ccb *container_controller.ContainerControlBlock) {
	ctr.resourceCtrl.RunningResourceRelease(ccb.CoreAllocated, ccb.MemoryRequest)
	ccb.UpdateRunningResourceAllocated([]string{})
}

func (ctr *DockerController) ContainerStart(ExtContainerID string) error {
	ccb, err := ctr.getCCB(ExtContainerID)
	if err != nil {
		return err
	}

	err = ctr.allocateRunningResourceForContainer(ccb)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			ctr.releaseRunningResourceForContainer(ccb)
		}
	}()

	err = ctr.cli.ContainerStart(context.Background(), ccb.ContainerID, types.ContainerStartOptions{})
	if err != nil {
		return types2.ErrInternalError
	}

	return nil
}
