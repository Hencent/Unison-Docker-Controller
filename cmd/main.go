package main

import (
	"fmt"
)
import "Unison-Docker-Controller/pkg"

func main() {
	fmt.Println("UDC: Starting UDC...")

	sysInfo, _ := pkg.NewDockerController()

	fmt.Println(sysInfo.SysBaseInfo)
}
