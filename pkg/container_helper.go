package pkg

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"strconv"
)

func generateExportsForContainer(tcpList []int, udpList []int) (nat.PortSet, error) {
	exports := make(nat.PortSet)
	for tcp := range tcpList {
		port, err := nat.NewPort("tcp", strconv.Itoa(tcp))
		if err != nil {
			return nil, err
		}
		exports[port] = struct{}{}
	}
	for udp := range udpList {
		port, err := nat.NewPort("udp", strconv.Itoa(udp))
		if err != nil {
			return nil, err
		}
		exports[port] = struct{}{}
	}

	return exports, nil
}

func generateCoreListStringWithIntList(vList []int) string {
	resp := ""
	for v := range vList {
		resp += "," + strconv.Itoa(v)
	}

	return resp[1:]
}

func (ctr *DockerController) containerRequestResource(containerID string) error {
	containerCfg := ctr.CCB[containerID].Config

	coreList, errCore := ctr.SysResource.CoreRequest(containerCfg.CoreCnt)
	if errCore != nil {
		return errCore
	}

	errRam := ctr.SysResource.RamRequest(uint64(containerCfg.RamAmount))
	if errRam != nil {
		ctr.SysResource.CoreRelease(coreList)
		return errRam
	}

	// TODO 限制磁盘资源

	// TODO 显卡资源

	updateConfig := container.UpdateConfig{
		Resources: container.Resources{
			Memory:     containerCfg.RamAmount,
			CpusetCpus: generateCoreListStringWithIntList(coreList),
			Devices:    nil,
		},
	}

	_, errUpdate := ctr.cli.ContainerUpdate(context.Background(), containerID, updateConfig)
	if errUpdate != nil {
		ctr.SysResource.CoreRelease(coreList)
		ctr.SysResource.RamRelease(uint64(containerCfg.RamAmount))
		return errUpdate
	}

	ctr.CCB[containerID].UpdateResource(coreList, containerCfg.RamAmount)
	return nil
}

func (ctr *DockerController) containerReleaseResource(containerID string) {
	ccb := ctr.CCB[containerID]

	ctr.SysResource.CoreRelease(ccb.Resource.CoreList)
	ctr.SysResource.RamRelease(uint64(ccb.Resource.RamAmount))

	ccb.UpdateResource([]int{}, 0)
}

func (ctr *DockerController) containerUpdateStatus(containerID string) {
	status := ctr.ContainerGetStatus(containerID)
	ctr.CCB[containerID].Status = status
}

func calculateCPUPercentUnix(v *types.StatsJSON) float64 {
	// 参考自
	// https://github.com/docker/cli/blob/902e9fa22bb7f591132ea52f333e6804eb0d46b6/cli/command/container/stats_helpers.go#L166

	previousCPU := v.PreCPUStats.CPUUsage.TotalUsage
	previousSystem := v.PreCPUStats.SystemUsage
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
		onlineCPUs  = float64(v.CPUStats.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}
