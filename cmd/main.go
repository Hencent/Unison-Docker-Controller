package main

import (
	"Unison-Docker-Controller/internal/config"
	"fmt"
)
import "Unison-Docker-Controller/pkg"

func main() {
	fmt.Println("UDC: Starting UDC...")

	sysInfo, _ := pkg.NewDockerController(config.Config{
		DockerContainerPath: "/home/penguincat/sundry",
	})

	fmt.Println(sysInfo)
	fmt.Println(sysInfo.SysBaseInfo)
	fmt.Println(sysInfo.SysDynamicInfo)
}
