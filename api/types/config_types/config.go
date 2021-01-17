package config_types

type Config struct {
	DockerContainerPath string
	//docker 容器存储路径下，保留的磁盘百分比 (0-100)
	DiskReserveRatio uint64

	//为系统运行保存的内存百分比 (0-100)
	RamReserveRatio uint64

	// container
	ContainerStopTimeout int
}
