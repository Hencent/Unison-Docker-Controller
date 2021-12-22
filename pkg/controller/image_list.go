package controller

import (
	"context"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/image"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"time"
)

func (ctr *DockerController) ImageList() ([]image.ImageListItem, error) {
	imagesSummary, err := ctr.cli.ImageList(context.Background(), types.ImageListOptions{
		All:     true,
		Filters: filters.Args{},
	})

	if err != nil {
		return nil, err
	}

	var images []image.ImageListItem
	for _, img := range imagesSummary {
		timeStr := time.Unix(img.Created, 0).Format("2006-01-02")
		images = append(images, image.ImageListItem{
			Name:        img.RepoTags[0],
			CreatedTime: timeStr,
			Size:        img.Size,
		})
	}

	return images, nil
}
