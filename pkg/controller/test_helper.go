package controller

import (
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/docker_controller"
	"testing"
)

func getDockerControllerForTest(t *testing.T) *DockerController {
	t.Helper()
	dc, _ := NewDockerController(&docker_controller.DockerControllerCreatBody{
		MemoryReserveRatio: 5,
		CoreAvailableList:  []string{"2", "3", "4", "5"},
		StoragePoolName:    "docker-thinpool",
		HostPortRange:      "14000-15000",
		Reload:             true,
	})

	return dc
}
