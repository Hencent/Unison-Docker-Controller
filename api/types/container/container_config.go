package container

type ContainerConfig struct {
	ImageName string

	ExposedTCPPorts []int
	ExposedUDPPorts []int

	Volumes []string

	CoreCnt    int
	RamAmount  uint64
	DiskAmount uint64

	ContainerName string
}
