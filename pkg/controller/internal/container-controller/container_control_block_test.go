package container_controller

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestContainerControlBlock_serialize(t *testing.T) {
	ccb := ContainerControlBlock{
		UECContainerID:         "123",
		ContainerID:            "456",
		ImageName:              "abc",
		ExposedTCPPorts:        []string{"1005", "1006"},
		ExposedTCPMappingPorts: []string{"1007", "1008"},
		ExposedUDPPorts:        []string{"1015", "1016"},
		ExposedUDPMappingPorts: []string{"1017", "1018"},
		CoreRequest:            4,
		CoreAllocated:          nil,
		MemoryRequest:          12345,
		StorageRequest:         54321,
		mu:                     sync.RWMutex{},
	}

	bytes, err := json.Marshal(&ccb)
	if err != nil {
		t.Fatal(err.Error())
	}

	var ccb2 ContainerControlBlock
	err = json.Unmarshal(bytes, &ccb2)
	if err != nil {
		t.Fatal(err.Error())
	}

	if ccb.UECContainerID != ccb2.UECContainerID || ccb.ExposedTCPPorts[0] != ccb2.ExposedTCPPorts[0] ||
		ccb.ExposedTCPPorts[1] != ccb2.ExposedTCPPorts[1] {
		t.Fatal("error")
	}
}
