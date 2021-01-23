package container_types

type ContainerStats struct {
	Memory uint64
	CPU    float64

	// TODO disk usage
	Disk uint64
}
