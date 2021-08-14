package controller

import (
	"context"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
	"github.com/docker/docker/api/types"
	"testing"
)

func TestContainerRemove(t *testing.T) {
	dc := getDockerControllerForTest(t)

	containerID, err := dc.ContainerCreate(container.ContainerCreateBody{
		ImageName:       "fedora",
		ExposedTCPPorts: []string{"1001", "1002"},
		ExposedUDPPorts: []string{"1003", "1004"},
		Mounts:          nil,
		CoreCnt:         2,
		MemorySize:      524288000,
		StorageSize:     21474836480,
	})
	if err != nil {
		t.Fatalf("unexpected error [%s]", err.Error())
		return
	}

	err = dc.ContainerRemove(containerID)

	if err != nil {
		_ = dc.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		t.Fatalf("remove container fail with message [%s]", err.Error())
	}
}
