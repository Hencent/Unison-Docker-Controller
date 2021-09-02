package container

type Stats uint

const (
	Running Stats = iota
	Creating
	Created
	Restarting
	Removing
	Stopping
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

func GetStatsString(stats Stats) string {
	switch stats {
	case Running:
		return "running"
	case Created:
		return "created"
	case Creating:
		return "creating"
	case Restarting:
		return "restarting"
	case Removing:
		return "removing"
	case Stopping:
		return "stopping"
	case Exited:
		return "exited"
	default:
		return "error"
	}
}
