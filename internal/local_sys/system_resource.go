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

var coreDistributeLock sync.Mutex
var ramDistributeLock sync.Mutex

type SystemResourceController struct {
	local_sys_types.SystemResource
	DockerContainerPath string
	//docker 容器存储路径下，保留的磁盘百分比 (0-100)
	DiskReserveRatio uint64
	//为系统运行保存的内存百分比 (0-100)
	RamReserveRatio uint64
}

func NewSystemResourceController(cfg config_types.Config, totalCoreCnt int, dockerRootDir string) (*SystemResourceController, error) {
	//ramLimit, errRam := getRamLimit(cfg.RamReserveRatio)
	//if errRam != nil {
	//	return nil, errRam
	//}
	//
	//diskLimit, errDisk := getDiskLimit(dockerRootDir, cfg.DiskReserveRatio)
	//if errDisk != nil {
	//	return nil, errDisk
	//}

	sysResource := &SystemResourceController{
		local_sys_types.SystemResource{
			RamAllocated:  0,
			DiskAllocated: 0,
		},
		dockerRootDir,
		cfg.DiskReserveRatio,
		cfg.RamReserveRatio,
	}

	errResource := sysResource.UpdateResourceLimit(nil)
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

//func getRamLimit(ramReserveRatio uint64) (uint64, error) {
//	virtualMemory, errVM := mem.VirtualMemory()
//	if errVM != nil {
//		return 0, errVM
//	}
//
//	return virtualMemory.Total * (100 - ramReserveRatio) / 100, nil
//}
//
//func getDiskLimit(dockerContainerPath string, diskReserveRatio uint64) (uint64, error) {
//	diskInfo, errDI := disk.Usage(dockerContainerPath)
//	if errDI != nil {
//		return 0, errDI
//	}
//
//	return diskInfo.Total * (100 - diskReserveRatio) / 100, nil
//}
//
//
//func (SystemResource *SystemResourceController) getRamAvailable() uint64 {
//	virtualMemory, errVM := mem.VirtualMemory()
//	if errVM != nil {
//		return 0
//	}
//	return virtualMemory.Available
//}
//
//func (SystemResource *SystemResourceController) getDiskAvailable() uint64 {
//	diskInfo, errDI := disk.Usage(SystemResource.DockerContainerPath)
//	if errDI != nil {
//		return 0
//	}
//	return diskInfo.Free
//}

func (SystemResource *SystemResourceController) UpdateResourceLimit(usage []*container_types.ContainerStats) error {
	containerDynamicMemUsage := uint64(0)
	containerDynamicDiskUsage := uint64(0)

	for _, v := range usage {
		containerDynamicMemUsage += v.Memory
		containerDynamicDiskUsage += v.Disk
	}

	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return errVM
	}
	diskInfo, errDI := disk.Usage(SystemResource.DockerContainerPath)
	if errDI != nil {
		return errDI
	}

	if virtualMemory.Available+containerDynamicMemUsage < virtualMemory.Total*(100-SystemResource.RamReserveRatio) ||
		diskInfo.Free+containerDynamicDiskUsage < diskInfo.Total*(100-SystemResource.DiskReserveRatio)/100 {
		return errors.New("system resource error")
	}

	SystemResource.RamLimit = virtualMemory.Available + containerDynamicMemUsage -
		virtualMemory.Total*(100-SystemResource.RamReserveRatio)/100
	SystemResource.DiskLimit = diskInfo.Free + containerDynamicDiskUsage -
		diskInfo.Total*(100-SystemResource.DiskReserveRatio)/100

	return nil
}

// Tip: request 资源之前，应当执行 UpdateResourceLimit

func (SystemResource *SystemResourceController) CoreRequest(cnt int) ([]int, error) {
	coreDistributeLock.Lock()
	defer coreDistributeLock.Unlock()

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
	ramDistributeLock.Lock()
	defer ramDistributeLock.Unlock()

	if amount+SystemResource.RamAllocated > SystemResource.RamLimit {
		return errors.New("not enough memory")
	}

	SystemResource.RamAllocated += amount
	return nil
}

func (SystemResource *SystemResourceController) RamRelease(amount uint64) {
	ramDistributeLock.Lock()
	defer ramDistributeLock.Unlock()

	if amount > SystemResource.RamAllocated {
		SystemResource.RamAllocated = 0
	} else {
		SystemResource.RamAllocated -= amount
	}
}

func (SystemResource *SystemResourceController) GetResourceAvailable() *local_sys_types.SystemResourceAvailable {
	return &local_sys_types.SystemResourceAvailable{
		AvailableRam:     SystemResource.RamLimit - SystemResource.RamAllocated,
		AvailableDisk:    SystemResource.DiskLimit - SystemResource.DiskAllocated,
		AvailableCoreCnt: SystemResource.AvailableCoreCnt,
	}
}
