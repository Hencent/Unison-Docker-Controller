package main

import (
	"Unison-Docker-Controller/api/types/config"
	"Unison-Docker-Controller/api/types/container"
	"fmt"
)
import "Unison-Docker-Controller/pkg"

func main() {
	fmt.Println("UDC: Starting UDC...")

	ctr, _ := pkg.NewDockerController(config.Config{
		DockerContainerPath: "/home/penguincat/sundry",
		DiskReserve:         0,
		RamReserve:          5,
	})

	fmt.Println(ctr)
	fmt.Println(ctr.SysBaseInfo)
	fmt.Println(ctr.SysDynamicInfo)

	cID, err := ctr.ContainerCreat(container.ContainerConfig{
		ImageName:     "penguincat/env:PYTORCH1.6",
		CoreCnt:       0,
		RamAmount:     0,
		DiskAmount:    0,
		ContainerName: "pcat",
	})
	fmt.Println(cID, err)
}
