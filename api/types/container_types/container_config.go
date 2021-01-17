package container_types

type ContainerConfig struct {
	ImageName string

	ExposedTCPPorts []int
	ExposedUDPPorts []int

	Volumes []string

	CoreCnt int
	// max memory usage, in bytes
	RamAmount int64
	// max disk usage, by bytes
	DiskAmount int64

	ContainerName string
}
