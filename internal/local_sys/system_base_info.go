package local_sys

import (
	"Unison-Docker-Controller/api/types/config"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

type SystemBaseInfo struct {
	platform        string
	PlatformFamily  string
	PlatformVersion string

	totalRam uint64

	cpuModelName  string
	logicalCores  int
	physicalCores int

	totalDisk uint64
}

func NewSystemBaseInfo(cfg config.Config) (*SystemBaseInfo, error) {
	sysInfo := new(SystemBaseInfo)

	err := sysInfo.initSystemInfo(cfg)
	if err != nil {
		return nil, err
	}

	return sysInfo, nil
}

func (sysBaseInfo *SystemBaseInfo) initSystemInfo(cfg config.Config) error {
	hostInfo, errHost := host.Info()
	if errHost != nil {
		return errHost
	}
	sysBaseInfo.platform = hostInfo.Platform
	sysBaseInfo.PlatformFamily = hostInfo.PlatformFamily
	sysBaseInfo.PlatformVersion = hostInfo.PlatformVersion

	cpuInfo, errCPU := cpu.Info()
	if errCPU != nil {
		return errCPU
	}
	sysBaseInfo.cpuModelName = cpuInfo[0].ModelName
	physicalCores, errPC := cpu.Counts(false)

	if errPC != nil {
		return errPC
	}
	sysBaseInfo.physicalCores = physicalCores

	logicalCores, errLC := cpu.Counts(true)
	if errLC != nil {
		return errLC
	}
	sysBaseInfo.logicalCores = logicalCores

	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return errVM
	}
	sysBaseInfo.totalRam = virtualMemory.Total

	diskInfo, errDI := disk.Usage(cfg.DockerContainerPath)
	if errDI != nil {
		return errDI
	}
	sysBaseInfo.totalDisk = diskInfo.Total

	return nil
}
