package container

// ContainerProfile
// static container info
type ContainerProfile struct {
	ContainerID string
	ImageName   string

	ExposedTCPPorts        []string
	ExposedTCPMappingPorts []string
	ExposedUDPPorts        []string
	ExposedUDPMappingPorts []string

	CoreRequest    int
	MemoryRequest  int64
	StorageRequest int64
}
