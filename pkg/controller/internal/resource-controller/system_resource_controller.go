package resource_controller

import (
	"container/list"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types"
	"github.com/PenguinCats/Unison-Docker-Controller/api/types/resource"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"

	//"github.com/shirou/gopsutil/disk"
	//"github.com/shirou/gopsutil/mem"
	"sync"
)

type ResourceControllerCreatBody struct {
	MemoryLimit   int64
	StorageLimit  int64
	CoreList      []string
	HostPortRange string
}

type ResourceController struct {
	memoryLimit     int64
	memoryAllocated int64

	storageLimit     int64
	storageAllocated int64

	coreAvailableList *list.List

	portAvailableList *list.List

	mu sync.RWMutex
}

func splitPortRange(s string) ([]string, error) {
	portRange := strings.Split(s, "-")
	if len(portRange) != 2 {
		logrus.Warning("invalid host port")
		return nil, types.ErrInternalError
	}
	portBegin, err := strconv.Atoi(portRange[0])
	if err != nil {
		logrus.Warning("invalid host port")
		return nil, types.ErrInternalError
	}

	portEnd, err := strconv.Atoi(portRange[1])
	if err != nil {
		logrus.Warning("invalid host port")
		return nil, types.ErrInternalError
	}

	if portBegin > portEnd {
		logrus.Warning("invalid host port")
		return nil, types.ErrInternalError
	}

	var ports []string
	for i := portBegin; i <= portEnd; i += 1 {
		ports = append(ports, strconv.Itoa(i))
	}

	return ports, nil
}

func NewResourceController(rcc *ResourceControllerCreatBody) (*ResourceController, error) {
	r := &ResourceController{
		memoryLimit:      rcc.MemoryLimit,
		memoryAllocated:  0,
		storageLimit:     rcc.StorageLimit,
		storageAllocated: 0,
	}

	r.coreAvailableList = list.New()
	for _, core := range rcc.CoreList {
		r.coreAvailableList.PushBack(core)
	}

	r.portAvailableList = list.New()
	ports, err := splitPortRange(rcc.HostPortRange)
	if err != nil {
		return nil, err
	}
	for _, port := range ports {
		r.portAvailableList.PushBack(port)
	}

	return r, nil
}

func (rc *ResourceController) CoreRequest(cnt int) ([]string, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if cnt > rc.coreAvailableList.Len() {
		return nil, types.ErrInsufficientResource
	}

	var cores []string

	for cnt > 0 {
		coreIDInterface := rc.coreAvailableList.Remove(rc.coreAvailableList.Front())
		coreID, _ := coreIDInterface.(string)
		cores = append(cores, coreID)
		cnt -= 1
	}

	return cores, nil
}

func (rc *ResourceController) CoreRelease(coreList []string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	for _, core := range coreList {
		rc.coreAvailableList.PushBack(core)
	}
}

func (rc *ResourceController) MemoryRequest(size int64) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.memoryLimit-rc.memoryAllocated < size {
		return types.ErrInsufficientResource
	}

	rc.memoryAllocated += size
	return nil
}

func (rc *ResourceController) MemoryRelease(size int64) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.storageAllocated -= size
}

func (rc *ResourceController) StorageRelease(size int64) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.memoryAllocated -= size
}

func (rc *ResourceController) StorageRequest(size int64) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.storageLimit-rc.storageAllocated < size {
		return types.ErrInsufficientResource
	}

	rc.storageAllocated += size
	return nil
}

func (rc *ResourceController) PortRequest(cnt int) ([]string, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if cnt > rc.portAvailableList.Len() {
		return nil, types.ErrInsufficientResource
	}

	var ports []string
	for cnt > 0 {
		portInterface := rc.portAvailableList.Remove(rc.portAvailableList.Front())
		portID, _ := portInterface.(string)
		ports = append(ports, portID)
		cnt -= 1
	}

	return ports, nil
}

func (rc *ResourceController) PortRelease(ports []string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	for _, port := range ports {
		rc.portAvailableList.PushBack(port)
	}
}

func (rc *ResourceController) FixedResourceRequest(storageSize int64, portCnt int) ([]string, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.storageLimit-rc.storageAllocated < storageSize ||
		rc.portAvailableList.Len() < portCnt {
		return nil, types.ErrInsufficientResource
	}

	rc.storageAllocated += storageSize

	var portList []string
	for portCnt > 0 {
		portInterface := rc.portAvailableList.Remove(rc.portAvailableList.Front())
		portID, _ := portInterface.(string)
		portList = append(portList, portID)
		portCnt -= 1
	}

	return portList, nil
}

func (rc *ResourceController) FixedResourceRelease(storageSize int64, portList []string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.storageAllocated -= storageSize

	for _, port := range portList {
		rc.portAvailableList.PushBack(port)
	}
}

func (rc *ResourceController) RunningResourceRequest(coreCnt int, memorySize int64) ([]string, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.coreAvailableList.Len() < coreCnt ||
		rc.memoryLimit-rc.memoryAllocated < memorySize {
		return nil, types.ErrInsufficientResource
	}

	var coreList []string
	for coreCnt > 0 {
		coreIDInterface := rc.coreAvailableList.Remove(rc.coreAvailableList.Front())
		coreID, _ := coreIDInterface.(string)
		coreList = append(coreList, coreID)
		coreCnt -= 1
	}

	rc.memoryLimit += memorySize

	return coreList, nil
}

func (rc *ResourceController) RunningResourceRelease(coreList []string, memorySize int64) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	for _, core := range coreList {
		rc.coreAvailableList.PushBack(core)
	}

	rc.memoryAllocated -= memorySize
}

func (rc *ResourceController) GetResourceAvailable() *resource.ResourceAvailable {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	return &resource.ResourceAvailable{
		MemoryAvailable:  rc.memoryLimit - rc.memoryAllocated,
		StorageAvailable: rc.storageLimit - rc.storageAllocated,
		CoreAvailable:    rc.coreAvailableList.Len(),
	}
}
