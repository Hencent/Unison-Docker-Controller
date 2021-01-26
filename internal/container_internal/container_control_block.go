package container_internal

import "Unison-Docker-Controller/api/types/container_types"

type ContainerStatus uint

const (
	Created ContainerStatus = iota
	Running
	Paused
	Restarting
	Removing
	Exited
	Dead
	Error
)

type internalResourceAllocated struct {
	CoreList  []int
	RamAmount int64
}

type ContainerControlBlock struct {
	Status            ContainerStatus
	Config            container_types.ContainerConfig
	ResourceAllocated internalResourceAllocated
	ResourceUsage     container_types.ContainerResourceUsage
}

func (ccb *ContainerControlBlock) UpdateResourceAllocated(coreList []int, ramAmount int64) {
	ccb.ResourceAllocated.CoreList = coreList
	ccb.ResourceAllocated.RamAmount = ramAmount
}
