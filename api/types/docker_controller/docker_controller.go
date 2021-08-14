package docker_controller

type DockerControllerCreatBody struct {
	MemoryReserveRatio int64

	StorageReserveRatioForImage int64
	StoragePoolName             string

	CoreAvailableList []string

	HostIP        string
	HostPortRange string // "14000-15000"
	HostPortBias  int

	ContainerStopTimeout int
}
