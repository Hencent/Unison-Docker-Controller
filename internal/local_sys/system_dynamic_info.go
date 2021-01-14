package local_sys

import (
	"Unison-Docker-Controller/internal/config"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"time"
)

type SystemDynamicInfo struct {
	availableRam  uint64
	coreLoad      []float64
	availableDisk uint64
}

func NewSystemDynamicInfo(cfg config.Config) (*SystemDynamicInfo, error) {
	sysInfo := new(SystemDynamicInfo)

	err := sysInfo.UpdateStatus(cfg)
	if err != nil {
		return nil, err
	}

	return sysInfo, nil
}

func (sysDynamicInfo *SystemDynamicInfo) UpdateStatus(cfg config.Config) error {
	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return errVM
	}
	sysDynamicInfo.availableRam = virtualMemory.Available
	ramReserve := virtualMemory.Total * cfg.RamReserve / 100
	if sysDynamicInfo.availableRam > ramReserve {
		sysDynamicInfo.availableRam -= ramReserve
	} else {
		sysDynamicInfo.availableRam = 0
	}

	coreLoad, errCL := cpu.Percent(time.Millisecond, true)
	if errCL != nil {
		return errCL
	}
	sysDynamicInfo.coreLoad = coreLoad

	diskInfo, errDI := disk.Usage(cfg.DockerContainerPath)
	if errDI != nil {
		return errDI
	}
	sysDynamicInfo.availableDisk = diskInfo.Free
	diskReserve := diskInfo.Total * cfg.DiskReserve / 100
	if sysDynamicInfo.availableDisk > diskReserve {
		sysDynamicInfo.availableDisk -= diskReserve
	} else {
		sysDynamicInfo.availableDisk = 0
	}

	return nil
}
