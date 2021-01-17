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

type internalResource struct {
	CoreList  []int
	RamAmount int64
}

type ContainerControlBlock struct {
	Status   ContainerStatus
	Config   container_types.ContainerConfig
	Resource internalResource
}

func (ccb *ContainerControlBlock) UpdateResource(coreList []int, ramAmount int64) {
	ccb.Resource.CoreList = coreList
	ccb.Resource.RamAmount = ramAmount
}
