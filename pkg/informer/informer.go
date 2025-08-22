// file: pkg/informer/informer.go

package informer

import (
	"context"
	"sync"
	"time"

	ecsmv1 "github.com/fx147/ecsm-operator/pkg/apis/ecsm/v1"
	"github.com/fx147/ecsm-operator/pkg/registry"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// ResourceEventHandler 是一组由业务控制器提供的回调函数。
// 我们直接复用 client-go 的定义。
type ResourceEventHandler = cache.ResourceEventHandler

// Informer 监听 Registry 的变更，并调用事件处理器。
type Informer interface {
	// AddEventHandler 注册一个事件处理器。
	AddEventHandler(handler ResourceEventHandler)
	// Run 启动 Informer 的主循环。
	Run(stopCh <-chan struct{})
}

// informer 是 Informer 接口的具体实现。
type informer struct {
	registry     registry.Interface // 数据源
	resyncPeriod time.Duration

	// --- 我们的核心状态 ---
	versionCache sync.Map // 线程安全的 "key -> resourceVersion" 缓存

	// --- 事件分发 ---
	handlers    []ResourceEventHandler
	handlerLock sync.RWMutex
}

// NewInformer 创建一个新的 Informer 实例。
func NewInformer(reg registry.Interface, resyncPeriod time.Duration) Informer {
	// 创建一个新的 informer 实例并返回
	inf := &informer{
		registry:     reg,
		resyncPeriod: resyncPeriod,
		handlers:     make([]ResourceEventHandler, 0),
	}

	return inf
}

func (i *informer) AddEventHandler(handler ResourceEventHandler) {
	i.handlerLock.Lock()
	defer i.handlerLock.Unlock()
	i.handlers = append(i.handlers, handler)
}

// distribute 将一个事件分发给所有已注册的处理器。
func (i *informer) distribute(eventType registry.EventType, obj interface{}) {
	i.handlerLock.RLock()
	defer i.handlerLock.RUnlock()

	for _, handler := range i.handlers {
		switch eventType {
		case registry.Added:
			handler.OnAdd(obj, false)
		case registry.Modified:
			// 注意：我们无法提供 oldObj，这是一个已知的设计权衡。
			// 我们传递新对象作为 old 和 new。
			handler.OnUpdate(obj, obj)
		case registry.Deleted:
			handler.OnDelete(obj)
		}
	}
}

func (i *informer) Run(stopCh <-chan struct{}) {
	klog.Infof("Starting informer...")

	// 1. 启动事件监听 goroutine
	go i.watchLoop(stopCh)

	// 2. 启动周期性 resync goroutine
	// 我们使用 wait.Until 来确保它在 stopCh 关闭时能正确退出
	go wait.Until(i.resync, i.resyncPeriod, stopCh)

	// 等待 stopCh 关闭
	<-stopCh
	klog.Infof("Shutting down informer...")
}

// watchLoop 消费来自 Registry 的实时事件
func (i *informer) watchLoop(stopCh <-chan struct{}) {
	eventCh, cancel := i.registry.Subscribe()
	defer cancel()

	for {
		select {
		case event, ok := <-eventCh:
			if !ok { // channel closed
				klog.Warningf("Registry event channel closed, watchLoop is stopping.")
				return
			}
			i.processEvent(event)
		case <-stopCh:
			return
		}
	}
}

// processEvent 处理单个实时事件
func (i *informer) processEvent(event registry.Event) {
	key := event.Key
	newRV := event.ResourceVersion

	// 从缓存中加载旧版本
	oldRV, exists := i.versionCache.Load(key)

	// 如果事件类型是删除，我们直接处理并从缓存中移除
	if event.Type == registry.Deleted {
		if exists {
			i.versionCache.Delete(key)
			i.distribute(event.Type, event.Object)
		}
		return
	}

	// 对于 Add 和 Update，如果版本没有变化，则忽略
	if exists && oldRV.(string) == newRV {
		return
	}

	// 版本有变化或对象是全新的，更新缓存并通知 handler
	i.versionCache.Store(key, newRV)
	i.distribute(event.Type, event.Object)
}

// resync 是我们的“安全网”
func (i *informer) resync() {
	klog.V(4).Infof("Running informer resync...")

	// 1. 从 Registry 全量 List 所有对象和当前的全局版本
	//    我们先只为 Service 实现
	allServices, _, err := i.registry.ListAllServices(context.Background(), "") // 假设 "" 表示所有命名空间
	if err != nil {
		klog.Errorf("Failed to list services for resync: %v", err)
		return
	}

	newVersionMap := make(map[string]string)

	// 2a. 找出 Added 和 Updated
	for _, service := range allServices.Items {
		key, _ := cache.MetaNamespaceKeyFunc(&service)
		newRV := service.ResourceVersion
		newVersionMap[key] = newRV

		oldRV, exists := i.versionCache.Load(key)

		if !exists {
			// 新增
			i.distribute(registry.Added, &service)
		} else if newRV != oldRV.(string) {
			// 更新
			i.distribute(registry.Modified, &service)
		}
	}

	// 2b. 找出 Deleted
	i.versionCache.Range(func(key interface{}, value interface{}) bool {
		if _, exists := newVersionMap[key.(string)]; !exists {
			// 构造一个 "tombstone" 对象来传递删除信息
			// 最简单的方法是创建一个只包含 key 信息的空对象
			deletedObj := &ecsmv1.ECSMService{}
			namespace, name, _ := cache.SplitMetaNamespaceKey(key.(string))
			deletedObj.Namespace = namespace
			deletedObj.Name = name
			deletedObj.ResourceVersion = value.(string) // 传递最后的版本号

			i.distribute(registry.Deleted, deletedObj)
		}
		return true
	})

	// 3. 用新的版本快照，更新 versionCache
	i.versionCache.Range(func(key, value interface{}) bool {
		if _, ok := newVersionMap[key.(string)]; !ok {
			i.versionCache.Delete(key)
		}
		return true
	})
	for key, rv := range newVersionMap {
		i.versionCache.Store(key, rv)
	}

	klog.V(4).Infof("Informer resync complete.")
}
