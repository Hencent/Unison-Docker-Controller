package controller

import (
	"context"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
	"github.com/docker/docker/api/types"
	"testing"
)

func TestContainerStop(t *testing.T) {
	dc := getDockerControllerForTest(t)

	storageSize := int64(15) * 1024 * 1024 * 1024
	containerID, err := dc.ContainerCreate(container.ContainerCreateBody{
		ImageName:       "fedora",
		ExposedTCPPorts: []string{"1001", "1002"},
		ExposedUDPPorts: []string{"1003", "1004"},
		Mounts:          nil,
		CoreCnt:         2,
		MemorySize:      524288000,
		StorageSize:     storageSize,
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

	err = dc.ContainerStart(containerID)
	if err != nil {
		t.Fatalf("unexpected error [%s]", err.Error())
	}

	err = dc.ContainerStop(containerID)
	if err != nil {
		t.Fatal("stop container fail")
	}
}
