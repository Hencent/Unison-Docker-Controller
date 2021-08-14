package controller

import (
	"context"
	"github.com/docker/docker/api/types"
)

func (ctr *DockerController) getInfo() (types.Info, error) {
	info, err := ctr.cli.Info(context.Background())
	if err != nil {
		return types.Info{}, err
	}

	return info, nil
}

func parseInfoForMemoryTotal(info *types.Info) int64 {
	return info.MemTotal
}
