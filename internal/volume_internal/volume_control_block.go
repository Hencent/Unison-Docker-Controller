package volume_internal

import "Unison-Docker-Controller/api/types/volume_types"

type VolumeControlBlock struct {
	ResourceUsage volume_types.VolumeResourceUsage
}
