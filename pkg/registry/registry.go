// file: pkg/registry/registry.go

package registry

import (
	"context"
	"encoding/binary"
	"sync"

	ecsmv1 "github.com/fx147/ecsm-operator/pkg/apis/ecsm/v1"
	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

var (
	// _metadataBucketKey 是一个特殊的 bucket，用于存放 registry 的元数据。
	_metadataBucketKey = []byte("_metadata")
	// _globalResourceVersionKey 是存储全局版本号的 key。
	_globalResourceVersionKey = []byte("globalResourceVersion")
)

// 编译时检查
var _ Interface = &Registry{}

// Interface 是 Registry 业务逻辑层的接口。
// 它定义了所有上层组件（如 Informer, Controller）可以调用的方法。
type Interface interface {
	// Subscribe 订阅 Registry 的变更事件。
	Subscribe() (<-chan Event, func())

	// -- Service-specific methods --
	CreateService(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error)
	UpdateService(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error)
	UpdateServiceStatus(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error)
	GetService(ctx context.Context, namespace, name string) (*ecsmv1.ECSMService, error)
	ListAllServices(ctx context.Context, namespace string) (*ecsmv1.ECSMServiceList, string, error)
	DeleteService(ctx context.Context, namespace, name string) error

	// -- Node-specific methods (future) --
	// ...

	// -- Image-specific methods (future) --
	// ...
}

// Registry 是业务逻辑层，它使用一个 Store 接口来持久化数据，并广播变更事件。
type Registry struct {
	db *bolt.DB // 直接持有 bbolt DB 实例以使用其事务

	// --- 事件相关的字段 ---
	subs      map[int]chan Event // 存储所有订阅者的 channel
	nextSubID int
	subsLock  sync.RWMutex // 保护 subs 字段的锁
}

// NewRegistry 创建一个新的 Registry 实例。
// 它接收一个已经打开的 bbolt 数据库实例。
func NewRegistry(db *bolt.DB) (*Registry, error) {
	// 初始化元数据 bucket
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(_metadataBucketKey)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Registry{
		db:   db,
		subs: make(map[int]chan Event),
	}, nil
}

// Subscribe 允许一个 Informer 或其他组件订阅 Registry 的变更事件。
// 它返回一个用于接收事件的 channel 和一个用于取消订阅的函数。
func (r *Registry) Subscribe() (<-chan Event, func()) {
	r.subsLock.Lock()
	defer r.subsLock.Unlock()

	id := r.nextSubID
	r.nextSubID++

	ch := make(chan Event, 100) // 使用带缓冲的 channel
	r.subs[id] = ch

	cancelFunc := func() {
		r.subsLock.Lock()
		defer r.subsLock.Unlock()
		if ch, ok := r.subs[id]; ok {
			close(ch)
			delete(r.subs, id)
		}
	}

	return ch, cancelFunc
}

// publish 是一个内部方法，用于向所有订阅者广播一个事件。
func (r *Registry) publish(event Event) {
	r.subsLock.RLock()
	defer r.subsLock.RUnlock()

	for _, ch := range r.subs {
		select {
		case ch <- event:
			// 发送成功
		default:
			// Channel is full, discard event.
			// This is acceptable because the periodic resync will eventually
			// correct any inconsistencies caused by missed events.
			klog.Warningf("Registry event channel is full. Discarding event for key %s.", event.Key)
		}
	}
}

// getAndIncrementGlobalRV 是一个在事务内部调用的辅助函数。
// 它原子性地获取并递增全局 resourceVersion。
// 这里为什么是原子性的？
func getAndIncrementGlobalRV(metaBucket *bolt.Bucket) (uint64, error) {
	currentRVBytes := metaBucket.Get(_globalResourceVersionKey)
	var currentRV uint64 = 0
	if currentRVBytes != nil {
		currentRV = binary.BigEndian.Uint64(currentRVBytes)
	}

	newRV := currentRV + 1

	newRVBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newRVBytes, newRV)

	if err := metaBucket.Put(_globalResourceVersionKey, newRVBytes); err != nil {
		return 0, err
	}

	return newRV, nil
}
