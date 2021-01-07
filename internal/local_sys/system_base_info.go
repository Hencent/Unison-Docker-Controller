package local_sys

import (
	"Unison-Docker-Controller/internal/config"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"runtime"
)

type SystemBaseInfo struct {
	osName string
	osArch string

	totalRam uint64

	logicalCores  int
	physicalCores int

	totalDisk uint64
}

func NewSystemBaseInfo(cfg config.Config) (*SystemBaseInfo, error) {
	sysInfo := new(SystemBaseInfo)

	sysInfo.osName = runtime.GOOS
	sysInfo.osArch = runtime.GOARCH

	err := sysInfo.initSystemInfo(cfg)
	if err != nil {
		return nil, err
	}

	return sysInfo, nil
}

func (sysInfo *SystemBaseInfo) initSystemInfo(cfg config.Config) error {
	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return errVM
	}
	sysInfo.totalRam = virtualMemory.Total

	physicalCores, errPC := cpu.Counts(false)
	if errPC != nil {
		return errPC
	}
	sysInfo.physicalCores = physicalCores

	logicalCores, errLC := cpu.Counts(true)
	if errLC != nil {
		return errLC
	}
	sysInfo.logicalCores = logicalCores

	diskInfo, errDI := disk.Usage(cfg.DockerContainerPath)
	if errDI != nil {
		return errDI
	}
	sysInfo.totalDisk = diskInfo.Total

	return nil
}
