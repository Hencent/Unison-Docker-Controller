package local_sys_types

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

type SystemBaseInfo struct {
	Platform        string
	PlatformFamily  string
	PlatformVersion string

	TotalRam uint64

	CpuModelName  string
	LogicalCores  int
	PhysicalCores int

	TotalDisk uint64
}

func NewSystemBaseInfo(dockerRootDIr string) (*SystemBaseInfo, error) {
	sysInfo := new(SystemBaseInfo)

	err := sysInfo.initSystemInfo(dockerRootDIr)
	if err != nil {
		return nil, err
	}

	return sysInfo, nil
}

func (sysBaseInfo *SystemBaseInfo) initSystemInfo(dockerRootDIr string) error {
	hostInfo, errHost := host.Info()
	if errHost != nil {
		return errHost
	}
	sysBaseInfo.Platform = hostInfo.Platform
	sysBaseInfo.PlatformFamily = hostInfo.PlatformFamily
	sysBaseInfo.PlatformVersion = hostInfo.PlatformVersion

	cpuInfo, errCPU := cpu.Info()
	if errCPU != nil {
		return errCPU
	}
	sysBaseInfo.CpuModelName = cpuInfo[0].ModelName
	physicalCores, errPC := cpu.Counts(false)

	if errPC != nil {
		return errPC
	}
	sysBaseInfo.PhysicalCores = physicalCores

	logicalCores, errLC := cpu.Counts(true)
	if errLC != nil {
		return errLC
	}
	sysBaseInfo.LogicalCores = logicalCores

	virtualMemory, errVM := mem.VirtualMemory()
	if errVM != nil {
		return errVM
	}
	sysBaseInfo.TotalRam = virtualMemory.Total

	diskInfo, errDI := disk.Usage(dockerRootDIr)
	if errDI != nil {
		return errDI
	}
	sysBaseInfo.TotalDisk = diskInfo.Total

	return nil
}
