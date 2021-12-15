package container_controller

import "sync"

type ContainerControlBlock struct {
	UECContainerID string `json:"uec_container_id"`
	ContainerID    string `json:"container_id"`
	ImageName      string `json:"image_name"`

	// Networks
	ExposedTCPPorts        []string `json:"exposed_tcp_ports"`
	ExposedTCPMappingPorts []string `json:"exposed_tcp_mapping_ports"`
	ExposedUDPPorts        []string `json:"exposed_udp_ports"`
	ExposedUDPMappingPorts []string `json:"exposed_udp_mapping_ports"`

	// Resource
	CoreRequest    int `json:"core_request"`
	CoreAllocated  []string
	MemoryRequest  int64 `json:"memory_request"`
	StorageRequest int64 `json:"storage_request"`

	mu sync.RWMutex
}

func (ccb *ContainerControlBlock) RenewMutexAfterReload() {
	ccb.mu = sync.RWMutex{}
}

func (ccb *ContainerControlBlock) UpdateRunningResourceAllocated(coreAllocated []string) {
	ccb.mu.Lock()
	defer ccb.mu.Unlock()

	ccb.CoreAllocated = coreAllocated
}
