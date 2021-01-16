package local_sys

import (
	"Unison-Docker-Controller/api/types/config"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type SystemDynamicInfo struct {
	AvailableRam uint64
	//CoreLoad      []float64
	AvailableDisk uint64

	AvailableCore []int
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
	sysDynamicInfo.AvailableRam = virtualMemory.Available
	ramReserve := virtualMemory.Total * cfg.RamReserve / 100
	if sysDynamicInfo.AvailableRam > ramReserve {
		sysDynamicInfo.AvailableRam -= ramReserve
	} else {
		sysDynamicInfo.AvailableRam = 0
	}

	//coreLoad, errCL := cpu.Percent(time.Millisecond, true)
	//if errCL != nil {
	//	return errCL
	//}
	//sysDynamicInfo.CoreLoad = coreLoad

	diskInfo, errDI := disk.Usage(cfg.DockerContainerPath)
	if errDI != nil {
		return errDI
	}
	sysDynamicInfo.AvailableDisk = diskInfo.Free
	diskReserve := diskInfo.Total * cfg.DiskReserve / 100
	if sysDynamicInfo.AvailableDisk > diskReserve {
		sysDynamicInfo.AvailableDisk -= diskReserve
	} else {
		sysDynamicInfo.AvailableDisk = 0
	}

	return nil
}
