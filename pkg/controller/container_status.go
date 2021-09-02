package controller

import (
	"context"
	"encoding/json"
	types2 "github.com/PenguinCats/Unison-Docker-Controller/api/types"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
	container2 "github.com/PenguinCats/Unison-Docker-Controller/pkg/controller/internal/container-controller"
	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
)

func (ctr *DockerController) containerStatsHelper(ccb *container2.ContainerControlBlock) (container.ContainerStatus, error) {
	cj, err := ctr.getContainerJson(ccb.ContainerID)
	if err != nil {
		return container.ContainerStatus{}, err
	}

	resp, err := ctr.cli.ContainerStatsOneShot(context.Background(), ccb.ContainerID)
	if err != nil {
		logrus.Warning(err.Error())
		return container.ContainerStatus{}, types2.ErrInternalError
	}
	dec := json.NewDecoder(resp.Body)
	var stats *types.StatsJSON
	err = dec.Decode(&stats)
	if err != nil {
		logrus.Warning(err.Error())
		return container.ContainerStatus{}, types2.ErrInternalError
	}

	// container stats
	cs := container.ContainerStatus{}
	switch cj.State.Status {
	case "created":
		cs.Stats = container.Created
	case "running":
		cs.Stats = container.Running
	case "restarting":
		cs.Stats = container.Restarting
	case "removing":
		cs.Stats = container.Removing
	case "exited":
		cs.Stats = container.Exited
	default:
		cs.Stats = container.Error
	}

	// resourceCtrl
	cs.CPUPercent = calculateCPUPercentUnix(stats.PreCPUStats.CPUUsage.TotalUsage, stats.PreCPUStats.SystemUsage, stats)
	cs.MemorySize = calculateMemUsageUnixNoCache(stats.MemoryStats)
	cs.MemoryPercent = calculateMemPercentUnixNoCache(float64(stats.MemoryStats.Limit), cs.MemorySize)
	// https://docs.docker.com/storage/storagedriver/#container-size-on-disk
	// https://stackoverflow.com/questions/22156563/what-is-the-exact-difference-between-sizerootfs-and-sizerw-in-docker-containers
	if cj.SizeRootFs != nil {
		cs.StorageSize = *cj.SizeRootFs
	} else {
		cs.StorageSize = 0
	}

	return cs, nil
}

func (ctr *DockerController) ContainerAllStats() map[string]container.ContainerStatus {
	mp := make(map[string]container.ContainerStatus)
	for _, ccb := range ctr.containerCtrlBlk {
		cs, err := ctr.containerStatsHelper(ccb)
		if err == nil {
			mp[ccb.UECContainerID] = cs
		}
	}
	return mp
}

func (ctr *DockerController) ContainerStats(ExtContainerID string) (container.ContainerStatus, error) {
	ccb, err := ctr.getCCB(ExtContainerID)
	if err != nil {
		return container.ContainerStatus{}, err
	}

	return ctr.containerStatsHelper(ccb)
}

func (ctr *DockerController) getContainerJson(containerID string) (types.ContainerJSON, error) {
	cj, err := ctr.cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		logrus.Warning(err.Error())
		return types.ContainerJSON{}, types2.ErrInternalError
	}

	return cj, nil
}

// calculate the CPUPercent usage. Ref:
// https://github.com/docker/cli/blob/902e9fa22bb7f591132ea52f333e6804eb0d46b6/cli/command/container/stats_helpers.go#L166
func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire hosts between readings
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

// Ref:
// https://github.com/docker/cli/blob/902e9fa22bb7f591132ea52f333e6804eb0d46b6/cli/command/container/stats_helpers.go#L239
// calculateMemUsageUnixNoCache calculate memory usage of the container.
// Cache is intentionally excluded to avoid misinterpretation of the output.
//
// On cgroup v1 hosts, the result is `mem.Usage - mem.Stats["total_inactive_file"]` .
// On cgroup v2 hosts, the result is `mem.Usage - mem.Stats["inactive_file"] `.
//
// This definition is consistent with cadvisor and containerd/CRI.
// * https://github.com/google/cadvisor/commit/307d1b1cb320fef66fab02db749f07a459245451
// * https://github.com/containerd/cri/commit/6b8846cdf8b8c98c1d965313d66bc8489166059a
//
// On Docker 19.03 and older, the result was `mem.Usage - mem.Stats["cache"]`.
// See https://github.com/moby/moby/issues/40727 for the background.
func calculateMemUsageUnixNoCache(mem types.MemoryStats) float64 {
	// cgroup v1
	if v, isCgroup1 := mem.Stats["total_inactive_file"]; isCgroup1 && v < mem.Usage {
		return float64(mem.Usage - v)
	}
	// cgroup v2
	if v := mem.Stats["inactive_file"]; v < mem.Usage {
		return float64(mem.Usage - v)
	}
	return float64(mem.Usage)
}

// Ref:
// https://github.com/docker/cli/blob/902e9fa22bb7f591132ea52f333e6804eb0d46b6/cli/command/container/stats_helpers.go#L251
func calculateMemPercentUnixNoCache(limit float64, usedNoCache float64) float64 {
	// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
	// got any data from cgroup
	if limit != 0 {
		return usedNoCache / limit * 100.0
	}
	return 0
}
