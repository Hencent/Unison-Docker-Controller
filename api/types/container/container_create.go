package container

type ContainerCreateBody struct {
	ImageName string

	ExposedTCPPorts []string
	ExposedUDPPorts []string

	Mounts []string

	CoreCnt int
	// max memory usage, in bytes
	MemorySize int64
	// max storage usage size, in bytes
	StorageSize int64
}
