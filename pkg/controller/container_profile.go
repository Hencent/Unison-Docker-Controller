package controller

import (
	"fmt"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
)

func (ctr *DockerController) ContainerProfile(containerID string) (container.ContainerProfile, error) {
	if !ctr.ContainerIsExist(containerID) {
		return container.ContainerProfile{}, fmt.Errorf("container [%s] does not exist", containerID)
	}

	ccb, err := ctr.getCCB(containerID)
	if err != nil {
		return container.ContainerProfile{}, err
	}

	cp := container.ContainerProfile{
		ContainerID:            containerID,
		ImageName:              ccb.ImageName,
		ExposedTCPPorts:        ccb.ExposedTCPPorts,
		ExposedTCPMappingPorts: ccb.ExposedTCPMappingPorts,
		ExposedUDPPorts:        ccb.ExposedUDPPorts,
		ExposedUDPMappingPorts: ccb.ExposedUDPMappingPorts,
		CoreRequest:            ccb.CoreRequest,
		MemoryRequest:          ccb.MemoryRequest,
		StorageRequest:         ccb.StorageRequest,
	}
	return cp, nil
}
