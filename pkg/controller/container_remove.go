package controller

import (
	"context"
	types2 "github.com/PenguinCats/Unison-Docker-Controller/api/types"
	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func (ctr *DockerController) ContainerRemove(ExtContainerID string) error {
	ccb, err := ctr.getCCB(ExtContainerID)
	if err != nil {
		return err
	}

	err = ctr.cli.ContainerRemove(context.Background(), ccb.ContainerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		logrus.Warning(err.Error())
		return types2.ErrInternalError
	}

	ctr.containerCtrlBlkMutex.Lock()
	delete(ctr.containerCtrlBlk, ExtContainerID)
	ctr.containerCtrlBlkMutex.Unlock()

	// 至此 container 删除成功， 清理对应的 leveldb
	err = ctr.db.Delete([]byte(ExtContainerID), &opt.WriteOptions{
		Sync: true,
	})
	if err != nil {
		logrus.Warning(err.Error())
		return types2.ErrLevelDbError
	}
	return nil
}
