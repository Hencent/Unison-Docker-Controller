package hosts

import (
	"errors"
	types2 "github.com/PenguinCats/Unison-Docker-Controller/api/types"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/hosts"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
)

var hostError = errors.New("get host info fail")

func GetHostInfo() (hosts.HostInfo, error) {
	hostInfo := hosts.HostInfo{}

	hInfo, err := host.Info()
	if err != nil {
		logrus.Warning(hostError.Error())
		return hosts.HostInfo{}, types2.ErrInternalError
	}
	hostInfo.Platform = hInfo.Platform
	hostInfo.PlatformFamily = hInfo.PlatformFamily
	hostInfo.PlatformVersion = hInfo.PlatformVersion

	cpuInfo, err := cpu.Info()
	if err != nil {
		logrus.Warning(hostError.Error())
		return hosts.HostInfo{}, types2.ErrInternalError
	}
	hostInfo.CpuModelName = cpuInfo[0].ModelName
	physicalCoreCnt, err := cpu.Counts(false)
	if err != nil {
		logrus.Warning(hostError.Error())
		return hosts.HostInfo{}, types2.ErrInternalError
	}
	hostInfo.PhysicalCoreCnt = physicalCoreCnt
	logicalCoreCnt, err := cpu.Counts(true)
	if err != nil {
		logrus.Warning(hostError.Error())
		return hosts.HostInfo{}, types2.ErrInternalError
	}
	hostInfo.LogicalCoreCnt = logicalCoreCnt

	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		logrus.Warning(hostError.Error())
		return hosts.HostInfo{}, types2.ErrInternalError
	}
	hostInfo.MemoryTotalSize = virtualMemory.Total

	return hostInfo, nil
}
