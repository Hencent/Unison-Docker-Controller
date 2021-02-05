package pkg

import (
	"Unison-Docker-Controller/api/types/container_types"
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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

	ctr.CCB[containerID].UpdateResourceAllocated(coreList, containerCfg.RamAmount)
	return nil
}

func (ctr *DockerController) containerReleaseResource(containerID string) {
	ccb := ctr.CCB[containerID]

	ctr.SysResource.CoreRelease(ccb.ResourceAllocated.CoreList)
	ctr.SysResource.RamRelease(uint64(ccb.ResourceAllocated.RamAmount))

	ccb.UpdateResourceAllocated([]int{}, 0)
}

func (ctr *DockerController) containerUpdateStatus(containerID string) {
	status := ctr.ContainerGetStatus(containerID)
	ctr.CCB[containerID].Status = status
}

func (ctr *DockerController) containerGetSingleCPUMemUsage(containerID string) (mem uint64, cpu float64, err error) {
	// 参考自
	// https://github.com/docker/cli/blob/902e9fa22bb7f591132ea52f333e6804eb0d46b6/cli/command/container/stats_helpers.go#L116

	resp, err := ctr.cli.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return 0, 0, err
	}

	dec := json.NewDecoder(resp.Body)
	var v *types.StatsJSON
	errJSON := dec.Decode(&v)
	if errJSON != nil {
		return 0, 0, errJSON
	}

	mem = v.Stats.MemoryStats.Usage - v.Stats.MemoryStats.Stats["cache"]
	cpu = calculateCPUPercentUnix(v)
	return mem, cpu, nil
}

func containerList(cli *client.Client) ([]types.Container, error) {
	containerList, errList := cli.ContainerList(context.Background(), types.ContainerListOptions{
		// https://docs.docker.com/engine/api/v1.41/#tag/Container
		// https://docs.docker.com/storage/storagedriver/#container-size-on-disk
		// https://stackoverflow.com/questions/22156563/what-is-the-exact-difference-between-sizerootfs-and-sizerw-in-docker-containers
		Size: true,
		All:  true,
	})
	if errList != nil {
		return nil, errList
	}

	return containerList, nil
}

func containerStopAndRemoveAllForInit(cli *client.Client, isStop bool, isRemove bool) {
	// 停止和删除所有现存容器只是一个尝试，若是出错，当前（2020.2.5）来看并不是特别重要，可以忽略
	containerList, err := containerList(cli)

	if err != nil {
		return
	}

	if isStop {
		for _, item := range containerList {
			_ = cli.ContainerStop(context.Background(), item.ID, nil)

			if isRemove {
				_ = cli.ContainerRemove(context.Background(), item.ID, types.ContainerRemoveOptions{
					RemoveVolumes: true,
					Force:         true,
				})
			}
		}
	}

}

func (ctr *DockerController) containerUpdateAllResourceUsage() error {
	// 周期性被调度

	containerList, errList := containerList(ctr.cli)
	if errList != nil {
		return errList
	}

	for i := 0; i < len(containerList); i++ {
		mem, cpu, errStats := ctr.containerGetSingleCPUMemUsage(containerList[i].ID)
		if errStats != nil {
			return errStats
		}

		ctr.CCB[containerList[i].ID].ResourceUsage = container_types.ContainerResourceUsage{
			Memory: mem,
			CPU:    cpu,
			Disk:   containerList[i].SizeRootFs,
		}
	}
	return nil
}
