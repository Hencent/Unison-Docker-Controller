package pkg

import (
	"context"
	"github.com/docker/docker/client"
)

func getDockerRootDir(client *client.Client) (string, error) {
	info, err := client.Info(context.Background())
	if err != nil {
		return "", err
	}

	return info.DockerRootDir, nil
}

func (ctr *DockerController) updateDynamicResource() error {
	containerDynamicUsage, errStats := ctr.ContainerAllStats()
	if errStats != nil {
		return errStats
	}

	errResourceUpdate := ctr.SysResource.UpdateResourceLimit(containerDynamicUsage)
	if errResourceUpdate != nil {
		return errResourceUpdate
	}

	return nil
}
