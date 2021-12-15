package controller

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/docker_controller"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/hosts"
	hosts2 "github.com/PenguinCats/Unison-Docker-Controller/internal/hosts"
	container2 "github.com/PenguinCats/Unison-Docker-Controller/pkg/controller/internal/container-controller"
	"github.com/PenguinCats/Unison-Docker-Controller/pkg/controller/internal/resource-controller"
	types2 "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DockerController struct {
	hostInfo     hosts.HostInfo
	resourceCtrl *resource_controller.ResourceController

	containerCtrlBlk      map[string]*container2.ContainerControlBlock
	containerCtrlBlkMutex sync.RWMutex
	containerStopTimeout  time.Duration

	cli *client.Client

	db *leveldb.DB
}

func (ctr *DockerController) getCCB(ExtContainerID string) (*container2.ContainerControlBlock, error) {
	ctr.containerCtrlBlkMutex.RLock()
	defer ctr.containerCtrlBlkMutex.RUnlock()

	if ccb, ok := ctr.containerCtrlBlk[ExtContainerID]; ok {
		return ccb, nil
	}

	logrus.Warningf("container [%s] does not exist", ExtContainerID)
	return nil, types.ErrContainerNotExist
}

func NewDockerController(dccb *docker_controller.DockerControllerCreatBody) (*DockerController, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	hostInfo, err := hosts2.GetHostInfo()
	if err != nil {
		return nil, err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	memoryTotal := int64(memInfo.Total)

	storageTotal, err := getStorageSize(dccb.StoragePoolName)
	if err != nil {
		return nil, err
	}

	rc, err := resource_controller.NewResourceController(&resource_controller.ResourceControllerCreatBody{
		MemoryLimit:   memoryTotal * (100 - dccb.MemoryReserveRatio) / 100,
		StorageLimit:  storageTotal * (100 - dccb.StorageReserveRatioForImage) / 100,
		CoreList:      dccb.CoreAvailableList,
		HostPortRange: dccb.HostPortRange,
	})
	if err != nil {
		return nil, err
	}

	if !dccb.Reload {
		// 不恢复，则删除元数据并实际清空所有 container
		dbPath := "/var/opt/uec/docker-controller.db"
		exist, err := isPathExists(dbPath)
		if err != nil {
			return nil, err
		}
		if exist {
			e := os.RemoveAll(dbPath)
			if e != nil {
				logrus.Warning(e.Error())
				return nil, err
			}
		}

		err = removeAllContainer(dockerClient)
		if err != nil {
			return nil, err
		}
	}
	db, err := leveldb.OpenFile("/var/opt/uec/docker-controller.db", nil)
	if err != nil {
		return nil, err
	}
	if dccb.Reload {
		//恢复，先暂停所有容器
		err := stopAllContainer(dockerClient)
		if err != nil {
			return nil, err
		}
	}

	c := &DockerController{
		hostInfo:             hostInfo,
		resourceCtrl:         rc,
		containerCtrlBlk:     make(map[string]*container2.ContainerControlBlock),
		containerStopTimeout: time.Duration(dccb.ContainerStopTimeout) * time.Second,
		cli:                  dockerClient,
		db:                   db,
	}

	// 恢复，填充 container control block
	if dccb.Reload {
		iter := db.NewIterator(nil, nil)
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			var ccb container2.ContainerControlBlock
			err := json.Unmarshal(value, &ccb)
			if err != nil {
				return nil, err
			}
			ccb.RenewMutexAfterReload()

			err = c.resourceCtrl.StorageRequest(ccb.StorageRequest)
			if err != nil {
				logrus.Warning(err.Error())
				return nil, err
			}
			c.containerCtrlBlk[string(key)] = &ccb
		}
	}

	//c.beginPeriodicTask()

	return c, nil
}

func isPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getStorageSize(storagePoolName string) (int64, error) {
	columns := []string{"NAME", "SIZE"}
	output, err := exec.Command(
		"lsblk",
		"-b", // output size in bytes
		"-P", // output fields as key=value pairs
		"-o", strings.Join(columns, ","),
	).Output()
	if err != nil {
		logrus.Warning("get storage size fail")
		return 0, types.ErrInternalError
	}

	var pairsRE = regexp.MustCompile(`([A-Z:]+)=(?:"(.*?)")`)
	s := bufio.NewScanner(bytes.NewReader(output))
	for s.Scan() {
		pairs := pairsRE.FindAllStringSubmatch(s.Text(), -1)
		if len(pairs) != 2 || len(pairs[0]) != 3 || len(pairs[1]) != 3 {
			logrus.Warning("get storage size fail")
			return 0, types.ErrInternalError
		}

		name := pairs[0][2]
		if name == storagePoolName {
			sizeString := pairs[1][2]
			size, err := strconv.ParseInt(sizeString, 10, 64)
			if err != nil {
				logrus.Warning("get storage size fail")
				return 0, types.ErrInternalError
			}

			return size, nil
		}
	}
	logrus.Warning("get storage size fail")
	return 0, types.ErrInternalError
}

func removeAllContainer(client *client.Client) error {
	containers, err := client.ContainerList(context.Background(), types2.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	for _, c := range containers {
		err = client.ContainerRemove(context.Background(), c.ID, types2.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func stopAllContainer(client *client.Client) error {
	containers, err := client.ContainerList(context.Background(), types2.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	for _, c := range containers {
		err = client.ContainerStop(context.Background(), c.ID, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctr *DockerController) GetHostInfo() hosts.HostInfo {
	return ctr.hostInfo
}
