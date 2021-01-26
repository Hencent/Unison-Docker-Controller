package config_types

type Config struct {
	//为系统运行保存的内存百分比 (0-100)
	RamReserveRatio uint64

	// container
	ContainerStopTimeout int
}
