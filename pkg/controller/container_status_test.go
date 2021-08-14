package controller

import (
	"context"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
	"github.com/docker/docker/api/types"
	"testing"
)

func TestContainerStatus(t *testing.T) {
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

	stats, err := dc.ContainerStats(containerID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if stats.Stats != container.Created ||
		stats.StorageSize != 0 ||
		stats.MemorySize != 0 ||
		stats.MemoryPercent != 0 ||
		stats.CPUPercent != 0 {
		t.Fatalf("wrong status")
	}

	err = dc.ContainerStart(containerID)
	if err != nil {
		t.Fatalf("unexpected error [%s]", err.Error())
	}
	stats, err = dc.ContainerStats(containerID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if stats.Stats != container.Running ||
		stats.MemorySize == 0 ||
		stats.MemoryPercent == 0 ||
		stats.CPUPercent == 0 {
		t.Fatalf("wrong status")
	}

	err = dc.ContainerStop(containerID)
	if err != nil {
		t.Fatalf("unexpected error [%s]", err.Error())
	}
	stats, err = dc.ContainerStats(containerID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if stats.Stats != container.Exited ||
		stats.MemorySize != 0 ||
		stats.MemoryPercent != 0 ||
		stats.CPUPercent != 0 {
		t.Fatalf("wrong status")
	}
}
