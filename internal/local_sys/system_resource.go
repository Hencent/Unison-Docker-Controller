package local_sys

import (
	"Unison-Docker-Controller/api/types/config_types"
	"errors"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"sync"
)

var coreDistributeLock sync.Mutex
var ramDistributeLock sync.Mutex

type SystemResource struct {
	ramLimit     uint64
	ramAllocated uint64

	dockerContainerPath string
	diskLimit           uint64
	diskAllocated       uint64

	availableCore    []bool
	availableCoreCnt int
}

func NewSystemResource(cfg config_types.Config, totalCoreCnt int) (*SystemResource, error) {
	ramLimit, errRam := getRamLimit(cfg.RamReserveRatio)
	if errRam != nil {
		return nil, errRam
	}

	diskLimit, errDisk := getDiskLimit(cfg.DockerContainerPath, cfg.DiskReserveRatio)
	if errDisk != nil {
		return nil, errDisk
	}

	sysInfo := &SystemResource{
		ramLimit:            ramLimit,
		ramAllocated:        0,
		dockerContainerPath: cfg.DockerContainerPath,
		diskLimit:           diskLimit,
		diskAllocated:       0,
	}

	sysInfo.availableCoreCnt = totalCoreCnt
	sysInfo.availableCore = make([]bool, totalCoreCnt)
	for k := range sysInfo.availableCore {
		sysInfo.availableCore[k] = true
	}

	return sysInfo, nil
}

func getRamLimit(ramReserveRatio uint64) (uint64, error) {
	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return 0, errVM
	}

	return virtualMemory.Total * (100 - ramReserveRatio) / 100, nil
}

func getDiskLimit(dockerContainerPath string, diskReserveRatio uint64) (uint64, error) {
	diskInfo, errDI := disk.Usage(dockerContainerPath)
	if errDI != nil {
		return 0, errDI
	}

	return diskInfo.Total * (100 - diskReserveRatio) / 100, nil
}

func (SystemResource *SystemResource) getRamAvailable() uint64 {
	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return 0
	}
	return virtualMemory.Available
}

func (SystemResource *SystemResource) getDiskAvailable() uint64 {
	diskInfo, errDI := disk.Usage(SystemResource.dockerContainerPath)
	if errDI != nil {
		return 0
	}
	return diskInfo.Free
}

func (SystemResource *SystemResource) CoreRequest(cnt int) ([]int, error) {
	coreDistributeLock.Lock()
	defer coreDistributeLock.Unlock()

	if cnt > SystemResource.availableCoreCnt {
		return nil, errors.New("not enough empty cores")
	}

	cores := make([]int, cnt)

	disCnt := 0
	for k, v := range SystemResource.availableCore {
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

func (SystemResource *SystemResource) CoreRelease(cores []int) {
	coreDistributeLock.Lock()
	defer coreDistributeLock.Unlock()

	cnt := 0
	for v := range cores {
		if SystemResource.availableCore[v] == false {
			cnt++
			SystemResource.availableCore[v] = true
		}
	}

	SystemResource.availableCoreCnt += cnt
}

func (SystemResource *SystemResource) RamRequest(amount uint64) error {
	if amount > SystemResource.getRamAvailable() {
		return errors.New("not enough memory")
	}

	ramDistributeLock.Lock()
	defer ramDistributeLock.Unlock()

	if amount+SystemResource.ramAllocated > SystemResource.ramLimit {
		return errors.New("not enough memory")
	}

	SystemResource.ramAllocated += amount
	return nil
}

func (SystemResource *SystemResource) RamRelease(amount uint64) {
	ramDistributeLock.Lock()
	defer ramDistributeLock.Unlock()

	if amount > SystemResource.ramAllocated {
		SystemResource.ramAllocated = 0
	} else {
		SystemResource.ramAllocated -= amount
	}
}
