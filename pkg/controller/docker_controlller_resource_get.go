package controller

import "github.com/PenguinCats/Unison-Docker-Controller/api/types/resource"

func (ctr *DockerController) GetResourceAvailable() *resource.ResourceAvailable {
	return ctr.resourceCtrl.GetResourceAvailable()
}
