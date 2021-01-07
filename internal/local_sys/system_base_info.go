package local_sys

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"runtime"
)

type SystemBaseInfo struct {
	osName string
	osArch string

	totalRam uint64

	logicalCores  int
	physicalCores int
}

func NewLocalSystemInfo() (*SystemBaseInfo, error) {
	sysInfo := &SystemBaseInfo{
		osName: runtime.GOOS,
		osArch: runtime.GOARCH,
	}

	err := sysInfo.initSystemInfo()
	if err != nil {
		return nil, err
	}

	return sysInfo, nil
}

func (sysInfo *SystemBaseInfo) initSystemInfo() error {
	v, err := mem.VirtualMemory()

	if err != nil {
		return err
	}

	sysInfo.totalRam = v.Total

	sysInfo.physicalCores, err = cpu.Counts(false)
	if err != nil {
		return err
	}

	sysInfo.logicalCores, err = cpu.Counts(true)
	if err != nil {
		return err
	}

	return nil
}
