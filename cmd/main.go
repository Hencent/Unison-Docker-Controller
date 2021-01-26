package main

import (
	"Unison-Docker-Controller/api/types/config_types"
	"Unison-Docker-Controller/api/types/container_types"
	"fmt"
)
import "Unison-Docker-Controller/pkg"

func main() {
	fmt.Println("UDC: Starting UDC...")

	ctr, errController := pkg.NewDockerController(config_types.Config{
		RamReserveRatio:      2,
		ContainerStopTimeout: 5,
	})
	fmt.Println(errController)

	fmt.Println(ctr)
	fmt.Println(ctr.SysBaseInfo)
	fmt.Println(ctr.SysResource)

	volumes := []string{"zhangbinjie", "tdd"}
	_ = ctr.VolumeCreate("zhangbinjie")
	_ = ctr.VolumeCreate("tdd")

	cID, err1 := ctr.ContainerCreat(container_types.ContainerConfig{
		ImageName:     "penguincat/env:PYTORCH1.6",
		CoreCnt:       2,
		RamAmount:     524288000,
		ContainerName: "pcat",
		Volumes:       volumes,
	})
	fmt.Println(cID, err1)

	err2 := ctr.ContainerStart(cID)
	fmt.Println(err2)

	//cID := "0dd144293cc4"

	usage, _ := ctr.ContainerResourceUsage(cID)
	memUsage := usage.Memory / 1024 / 1024
	fmt.Println(memUsage)

	err3 := ctr.ContainerStop(cID)
	fmt.Println(err3)

	err4 := ctr.ContainerRemove(cID)
	fmt.Println(err4)

	_ = ctr.VolumeRemove("zhangbinjie", false)
	_ = ctr.VolumeRemove("tdd", false)
}
