package controller

import (
	"context"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
	"github.com/PenguinCats/Unison-Docker-Controller/internal/utils"
	"github.com/docker/docker/api/types"
	"testing"
)

func TestContainerProfile(t *testing.T) {
	dc := getDockerControllerForTest(t)

	storageSize := int64(15) * 1024 * 1024 * 1024
	ExtContainerID := "123456"
	containerID, err := dc.ContainerCreate(container.ContainerCreateBody{
		ImageName:       "fedora",
		ExposedTCPPorts: []string{"1001", "1002"},
		ExposedUDPPorts: []string{"1003", "1004"},
		Mounts:          nil,
		CoreCnt:         2,
		MemorySize:      524288000,
		StorageSize:     storageSize,
		ExtContainerID:  ExtContainerID,
	})
	if err != nil {
		t.Fatalf("unexpected error [%s]", err.Error())
	}

	defer func() {
		err := dc.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil {
			println(err.Error())
		}
	}()

	profile, err := dc.ContainerProfile(ExtContainerID)
	if err != nil {
		t.Fatalf("get container profile fail with [%s]", err.Error())
	}

	if profile.ImageName != "fedora" ||
		!utils.CompareSliceString(profile.ExposedTCPPorts, []string{"1001", "1002"}) ||
		!utils.CompareSliceString(profile.ExposedUDPPorts, []string{"1003", "1004"}) ||
		profile.CoreRequest != 2 ||
		profile.MemoryRequest != 524288000 ||
		profile.StorageRequest != storageSize {
		t.Fatal("get container profile fail with wrong information")
	}
}
