package local_sys_types

type SystemResource struct {
	RamLimit     uint64
	RamAllocated uint64

	DiskLimit     uint64
	DiskAllocated uint64

	AvailableCore    []bool
	AvailableCoreCnt int
}

type SystemResourceAvailable struct {
	AvailableRam     uint64
	AvailableDisk    uint64
	AvailableCoreCnt int
}
