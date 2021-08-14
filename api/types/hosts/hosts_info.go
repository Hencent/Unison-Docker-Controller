package hosts

type HostInfo struct {
	Platform        string
	PlatformFamily  string
	PlatformVersion string

	MemoryTotalSize uint64

	CpuModelName    string
	LogicalCoreCnt  int
	PhysicalCoreCnt int
}
