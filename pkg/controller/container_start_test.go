package controller

import (
	"context"
	"fmt"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
	"github.com/docker/docker/api/types"
	"testing"
)

func TestContainerStart(t *testing.T) {
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
		fmt.Println(err.Error())
		t.Fatalf("container [%s] start fail", containerID)
	}

	stats, err := dc.ContainerStats(containerID)
	if err != nil {
		t.Fatalf("unexpected error [%s]", err.Error())
	}

	if stats.Stats != container.Running {
		t.Fatalf("container [%s] start fail: is not running", containerID)
	}
}
