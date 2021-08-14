package container

type Stats uint

const (
	Running Stats = iota
	Created
	Restarting
	Removing
	Exited
	Error
)

type ContainerStatus struct {
	// container stats
	Stats

	// resource
	CPUPercent    float64
	MemoryPercent float64
	MemorySize    float64
	StorageSize   int64
}
