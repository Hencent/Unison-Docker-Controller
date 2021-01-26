package local_sys

import (
	"Unison-Docker-Controller/api/types/config_types"
	"Unison-Docker-Controller/api/types/container_types"
	"Unison-Docker-Controller/api/types/local_sys_types"
	"errors"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"sync"
)

var coreResourceLock sync.Mutex
var ramResourceLock sync.Mutex

type SystemResourceController struct {
	local_sys_types.SystemResource
	DockerContainerPath string
	//为系统运行保存的内存百分比 (0-100)
	RamReserveRatio uint64
}

func NewSystemResourceController(cfg config_types.Config, totalCoreCnt int, dockerRootDir string) (*SystemResourceController, error) {
	sysResource := &SystemResourceController{
		local_sys_types.SystemResource{
			RamAllocated: 0,
		},
		dockerRootDir,
		cfg.RamReserveRatio,
	}

	errResource := sysResource.UpdateResourceStates(nil)
	if errResource != nil {
		return nil, errResource
	}

	sysResource.AvailableCoreCnt = totalCoreCnt
	sysResource.AvailableCore = make([]bool, totalCoreCnt)
	for k := range sysResource.AvailableCore {
		sysResource.AvailableCore[k] = true
	}

	return sysResource, nil
}

func (SystemResource *SystemResourceController) UpdateResourceStates(usage map[string]*container_types.ContainerResourceUsage) error {
	containerDynamicMemUsage := uint64(0)

	for _, v := range usage {
		containerDynamicMemUsage += v.Memory
	}

	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return errVM
	}
	diskInfo, errDI := disk.Usage(SystemResource.DockerContainerPath)
	if errDI != nil {
		return errDI
	}

	if virtualMemory.Available < virtualMemory.Total*SystemResource.RamReserveRatio/100 {
		return errors.New("system resource exceeded")
	}

	ramResourceLock.Lock()
	SystemResource.RamLimit = virtualMemory.Available - virtualMemory.Total*SystemResource.RamReserveRatio/100 + containerDynamicMemUsage
	ramResourceLock.Unlock()

	SystemResource.AvailableDisk = diskInfo.Free

	return nil
}

// Tip: request 资源之前，应当执行 UpdateResourceStates

func (SystemResource *SystemResourceController) CoreRequest(cnt int) ([]int, error) {
	coreResourceLock.Lock()
	defer coreResourceLock.Unlock()

	if cnt > SystemResource.AvailableCoreCnt {
		return nil, errors.New("not enough empty cores")
	}

	cores := make([]int, cnt)

	disCnt := 0
	for k, v := range SystemResource.AvailableCore {
		if disCnt >= cnt {
			break
		}
		if v {
			cores[disCnt] = k
			disCnt++
		}
	}

	return cores, nil
}

func (SystemResource *SystemResourceController) CoreRelease(cores []int) {
	cnt := 0
	for v := range cores {
		if SystemResource.AvailableCore[v] == false {
			cnt++
			SystemResource.AvailableCore[v] = true
		}
	}

	SystemResource.AvailableCoreCnt += cnt
}

func (SystemResource *SystemResourceController) RamRequest(amount uint64) error {
	ramResourceLock.Lock()
	defer ramResourceLock.Unlock()

	if amount+SystemResource.RamAllocated > SystemResource.RamLimit {
		return errors.New("not enough memory")
	}

	SystemResource.RamAllocated += amount
	return nil
}

func (SystemResource *SystemResourceController) RamRelease(amount uint64) {
	ramResourceLock.Lock()
	defer ramResourceLock.Unlock()

	if amount > SystemResource.RamAllocated {
		SystemResource.RamAllocated = 0
	} else {
		SystemResource.RamAllocated -= amount
	}
}

func (SystemResource *SystemResourceController) GetResourceAvailable() *local_sys_types.SystemResourceAvailable {
	return &local_sys_types.SystemResourceAvailable{
		AvailableRam:     SystemResource.RamLimit - SystemResource.RamAllocated,
		AvailableDisk:    SystemResource.AvailableDisk,
		AvailableCoreCnt: SystemResource.AvailableCoreCnt,
	}
}
