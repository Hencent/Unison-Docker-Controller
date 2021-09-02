package controller

import (
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
)

func (ctr *DockerController) ContainerProfile(ExtContainerID string) (container.ContainerProfile, error) {
	ccb, err := ctr.getCCB(ExtContainerID)
	if err != nil {
		return container.ContainerProfile{}, err
	}

	cp := container.ContainerProfile{
		ExtContainerID:         ExtContainerID,
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
