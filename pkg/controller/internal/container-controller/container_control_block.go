package container_controller

import "sync"

type ContainerControlBlock struct {
	ContainerID string
	ImageName   string

	// Networks
	ExposedTCPPorts        []string
	ExposedTCPMappingPorts []string
	ExposedUDPPorts        []string
	ExposedUDPMappingPorts []string

	// Resource
	CoreRequest    int
	CoreAllocated  []string
	MemoryRequest  int64
	StorageRequest int64

	mu sync.RWMutex
}

func (ccb *ContainerControlBlock) UpdateRunningResourceAllocated(coreAllocated []string) {
	ccb.mu.Lock()
	defer ccb.mu.Unlock()

	ccb.CoreAllocated = coreAllocated
}
