package docker_controller

type DockerControllerCreatBody struct {
	MemoryReserveRatio int64

	StorageReserveRatioForImage int64
	StoragePoolName             string

	CoreAvailableList []string

	HostPortRange string // "14000-15000"

	ContainerStopTimeout int

	Reload bool
}
