package controller

import (
	"context"
	"encoding/json"
	types2 "github.com/PenguinCats/Unison-Docker-Controller/api/types"
	container2 "github.com/PenguinCats/Unison-Docker-Controller/api/types/container"
	"github.com/PenguinCats/Unison-Docker-Controller/pkg/controller/internal/container-controller"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strconv"
)

func generatePortBinding(tcpList, tcpBindingList, udpList, udpBindingList []string) (nat.PortSet, nat.PortMap, error) {
	portBinding := nat.PortMap{}
	portSet := nat.PortSet{}
	for idx := range tcpList {
		port, err := nat.NewPort("tcp", tcpList[idx])
		if err != nil {
			logrus.Warning(err.Error())
			return nil, nil, types2.ErrInternalError
		}
		portBinding[port] = []nat.PortBinding{
			{
				HostPort: tcpBindingList[idx],
			},
		}
		portSet[port] = struct{}{}
	}
	for idx := range udpList {
		port, err := nat.NewPort("udp", udpList[idx])
		if err != nil {
			logrus.Warning(err.Error())
			return nil, nil, types2.ErrInternalError
		}
		portBinding[port] = []nat.PortBinding{
			{
				HostPort: udpBindingList[idx],
			},
		}
		portSet[port] = struct{}{}
	}

	return portSet, portBinding, nil
}

func (ctr *DockerController) ContainerCreate(cb container2.ContainerCreateBody) (string, error) {
	//mountInfo := make([]mount.Mount, len(cb.Mounts))
	//for k, v := range cb.Mounts {
	//	mountInfo[k] = mount.Mount{
	//		Type:   "volume-controller",
	//		Source: v,
	//		Target: path.Join("/volume-controller", v),
	//	}
	//}

	storageSizeInt64 := cb.StorageSize / 1073741824 * 1073741824
	storageSizeG := strconv.FormatInt(cb.StorageSize/1073741824, 10)
	storageOpt := map[string]string{
		"size": storageSizeG + "G",
	}

	portCnt := len(cb.ExposedTCPPorts) + len(cb.ExposedUDPPorts)
	portList, err := ctr.resourceCtrl.FixedResourceRequest(storageSizeInt64, portCnt)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			ctr.resourceCtrl.FixedResourceRelease(storageSizeInt64, portList)
		}
	}()

	tcpPortBinding, udpPortBinding := portList[:len(cb.ExposedTCPPorts)], portList[len(cb.ExposedTCPPorts):]
	portSet, portMapping, err := generatePortBinding(cb.ExposedTCPPorts, tcpPortBinding, cb.ExposedUDPPorts, udpPortBinding)
	if err != nil {
		return "", err
	}

	resp, err := ctr.cli.ContainerCreate(context.Background(),
		&container.Config{
			Image:        cb.ImageName,
			Tty:          true,
			ExposedPorts: portSet,
		}, &container.HostConfig{
			//Mounts:     mountInfo,
			StorageOpt: storageOpt,
			Resources: container.Resources{
				Memory: cb.MemorySize,
			},
			PortBindings: portMapping,
		}, nil, nil, cb.ExtContainerID)

	if err != nil {
		logrus.Warning(err.Error())
		return "", types2.ErrInternalError
	}

	defer func() {
		if err != nil {
			_ = ctr.cli.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			})
		}
	}()

	ccb := &container_controller.ContainerControlBlock{
		UECContainerID:         cb.ExtContainerID,
		ContainerID:            resp.ID,
		ImageName:              cb.ImageName,
		ExposedTCPPorts:        cb.ExposedTCPPorts,
		ExposedTCPMappingPorts: tcpPortBinding,
		ExposedUDPPorts:        cb.ExposedUDPPorts,
		ExposedUDPMappingPorts: udpPortBinding,
		CoreRequest:            cb.CoreCnt,
		MemoryRequest:          cb.MemorySize,
		StorageRequest:         storageSizeInt64,
	}
	ctr.containerCtrlBlkMutex.Lock()
	ctr.containerCtrlBlk[cb.ExtContainerID] = ccb
	ctr.containerCtrlBlkMutex.Unlock()

	// 至此 container 创建成功，写入 leveldb 中以备日后恢复
	ccbByte, err := json.Marshal(&ccb)
	if err != nil {
		return "", err
	}
	err = ctr.db.Put([]byte(cb.ExtContainerID), ccbByte, &opt.WriteOptions{
		Sync: true,
	})
	if err != nil {
		logrus.Warning(err.Error())
		return "", types2.ErrLevelDbError
	}

	return resp.ID, nil
}
