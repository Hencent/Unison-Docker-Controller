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

func (sysInfo *SystemDynamicInfo) UpdateStatus(cfg config.Config) error {
	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return errVM
	}

	sysInfo.availableRam = virtualMemory.Available

	coreLoad, errCL := cpu.Percent(time.Millisecond, true)
	if errCL != nil {
		return errCL
	}

	sysInfo.coreLoad = coreLoad

	diskInfo, errDI := disk.Usage(cfg.DockerContainerPath)
	if errDI != nil {
		return errDI
	}
	sysInfo.availableDisk = diskInfo.Free

	return nil
}
