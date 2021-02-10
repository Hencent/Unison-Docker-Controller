package config_types

type Config struct {
	//为系统运行保存的内存百分比 (0-100)
	RamReserveRatio uint64

	// container
	ContainerStopTimeout int

	// 周期性任务执行间隔 by second
	PeriodicSystemStatsUpdateInterval int

	// 启动时对现存容器的操作
	StopExistingContainersOnStart   bool
	RemoveExistingContainersOnStart bool

	// 启动时对现存 Volume 的操作
	RemoveExistingVolumesOnStart bool
}
