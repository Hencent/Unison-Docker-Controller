package pkg

import (
	"Unison-Docker-Controller/api/types/volume_types"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func volumeList(cli *client.Client) ([]*types.Volume, error) {
	containerListBody, errList := cli.VolumeList(context.Background(), filters.Args{})
	if errList != nil {
		return nil, errList
	}

	return containerListBody.Volumes, nil
}

func (ctr *DockerController) volumeUpdateAllResourceUsage() error {
	// 周期性被调度

	volumeList, errList := volumeList(ctr.cli)
	if errList != nil {
		return errList
	}

	for i := 0; i < len(volumeList); i++ {
		usage, errUsage := ctr.VolumeResourceUsage(volumeList[i].Name)
		if errUsage != nil {
			return errUsage
		}

		ctr.VCB[volumeList[i].Name].ResourceUsage = volume_types.VolumeResourceUsage{
			RefCount: usage.RefCount,
			Size:     usage.Size,
		}
	}
	return nil
}
