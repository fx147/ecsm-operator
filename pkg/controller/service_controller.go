// file: pkg/controller/service_controller.go

package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	ecsmv1 "github.com/fx147/ecsm-operator/pkg/apis/ecsm/v1"
	"github.com/fx147/ecsm-operator/pkg/ecsm-client/clientset"
	"github.com/fx147/ecsm-operator/pkg/informer"
	"github.com/fx147/ecsm-operator/pkg/registry"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	// maxRetries 是一个 key 在被放弃前的最大重试次数。
	maxRetries = 15
)

// ECSMServiceController 负责监听 ECSMService 对象的变更，
// 并确保 ECSM 平台上的真实状态与对象的 spec 保持一致。
type ECSMServiceController struct {
	// clientset 用于与 ECSM API Server 交互 (现实世界)
	ecsmClient clientset.Interface

	// registry 用于更新我们自己存储中的对象状态 (期望世界)
	registry registry.Interface

	// serviceLister 提供了从 Informer 缓存中读取 ECSMService 的能力。
	// 注意：我们需要自己实现一个 Lister，或者直接使用 Informer 的缓存。
	// 为了简化，我们先假设 Informer 提供了 Get 方法。
	serviceInformer informer.Informer // 我们自己的 Informer

	// queue 是一个限速工作队列。
	queue workqueue.TypedRateLimitingInterface[interface{}]
}

// NewECSMServiceController 创建一个新的控制器实例。
func NewECSMServiceController(
	ecsmClient clientset.Interface,
	reg registry.Interface,
	serviceInformer informer.Informer,
) *ECSMServiceController {

	c := &ECSMServiceController{
		ecsmClient:      ecsmClient,
		registry:        reg,
		serviceInformer: serviceInformer,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ecsmservice"),
	}

	// EventHandler 的唯一职责就是将事件的 key 推入队列。
	// 它不关心对象内容。
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, _ := cache.MetaNamespaceKeyFunc(obj)
			c.queue.Add(key)
		},
		UpdateFunc: func(old, new interface{}) {
			key, _ := cache.MetaNamespaceKeyFunc(new)
			c.queue.Add(key)
		},
		DeleteFunc: func(obj interface{}) {
			key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			c.queue.Add(key)
		},
	}

	serviceInformer.AddEventHandler(handler)

	return c
}

// enqueueService 将一个 ECSMService 的 key 添加到工作队列中。
func (c *ECSMServiceController) enqueueService(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.queue.Add(key)
}

// Run 启动控制器的主工作循环。
func (c *ECSMServiceController) Run(workers int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting ECSMService controller")
	defer klog.Info("Shutting down ECSMService controller")

	// 启动 Informer，它会开始填充缓存和监听事件
	// 注意：Informer 应该在控制器外部被启动和管理
	// 我们假设调用 Run 的地方已经启动了 Informer

	klog.Info("Waiting for informer caches to sync...")
	// if !cache.WaitForCacheSync(stopCh, c.serviceInformer.HasSynced) {
	// 	runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	// 	return
	// }
	// (在我们的模型中，我们没有 HasSynced，所以暂时注释掉)

	klog.Info("Starting workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

// runWorker 是一个持续运行的循环，负责从队列中消费任务并处理。
func (c *ECSMServiceController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem 从队列中取出一个任务，并调用 reconcile 来处理它。
func (c *ECSMServiceController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.reconcile(key.(string))
	// 调用我们之前在 K8s 中看到的 handleErr 逻辑
	c.handleErr(err, key)

	return true
}

// handleErr 负责处理 reconcile 返回的错误，并决定是否重试。
func (c *ECSMServiceController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		klog.V(2).Infof("Error syncing service %v: %v. Retrying.", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	runtime.HandleError(err)
	klog.Warningf("Dropping service %q out of the queue: %v", key, err)
	c.queue.Forget(key)
}

func (c *ECSMServiceController) reconcile(key string) error {
	klog.Infof("Reconciling ECSMService %s", key)
	ctx := context.Background()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		// This is a programming error, a malformed key was put into the queue.
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil // We don't requeue programming errors.
	}

	// --- 1. 从 Registry 获取“期望” (`Spec`) ---
	//    这是我们架构的核心：直接访问持久化层。
	desiredService, err := c.registry.GetService(ctx, namespace, name)
	if err != nil {
		if errors.IsNotFound(err) {
			// 对象已被删除，无需处理。Informer 的 resync 会清理 versionCache。
			klog.Infof("ECSMService %s in work queue no longer exists", key)
			return nil
		}
		return err // 其他读取错误，需要重试
	}

	// --- 2. 获取“现实” ---
	//    调用 EcsmClient
	actualContainers, err := c.ecsmClient.Containers().ListAllByService(ctx, clientset.ListContainersByServiceOptions{
		ServiceIDs: []string{string(desiredService.UID)},
	})
	if err != nil {
		// 如果是网络错误等，返回 err 会触发重试
		return fmt.Errorf("failed to list containers for service %s: %w", key, err)
	}

	// --- 3. 调谐 (Compare & Act) ---
	desiredReplicas := 0
	if desiredService.Spec.DeploymentStrategy.Replicas != nil {
		desiredReplicas = int(*desiredService.Spec.DeploymentStrategy.Replicas)
	}
	actualReplicas := len(actualContainers)

	delta := desiredReplicas - actualReplicas

	if delta > 0 {
		klog.Infof("Service %s: Desired replicas (%d) > Actual (%d). Need to create %d container(s).", key, desiredReplicas, actualReplicas, delta)
		// TODO: 在这里实现创建容器的逻辑
		// err := c.createContainers(ctx, delta, desiredService)
		// return err
	} else if delta < 0 {
		klog.Infof("Service %s: Desired replicas (%d) < Actual (%d). Need to delete %d container(s).", key, desiredReplicas, actualReplicas, -delta)
		// TODO: 在这里实现删除容器的逻辑
		// err := c.deleteContainers(ctx, -delta, actualContainers)
		// return err
	}

	// TODO: 在这里实现滚动更新的逻辑，比较 template spec 和容器的 image/config

	// --- 4. 更新“状态” (`Status`) ---
	// 重新获取最新的现实快照，因为我们可能刚刚修改了它
	finalContainers, err := c.ecsmClient.Containers().ListAllByService(ctx, clientset.ListContainersByServiceOptions{
		ServiceIDs: []string{string(desiredService.UID)},
	})
	if err != nil {
		return fmt.Errorf("failed to list containers for status update for service %s: %w", key, err)
	}

	newStatus := c.calculateStatus(finalContainers)

	// 只有当 status 真的变了，才去写 Registry
	if !reflect.DeepEqual(desiredService.Status, newStatus) {
		klog.Infof("Updating status for service %s", key)
		serviceToUpdate := desiredService.DeepCopy()
		serviceToUpdate.Status = newStatus
		// 注意：这里我们应该使用 UpdateServiceStatus，而不是 UpdateService
		// 以防止覆盖用户可能同时对 spec 做的修改
		_, err := c.registry.UpdateServiceStatus(ctx, serviceToUpdate)
		return err // 返回错误以触发可能的重试
	}

	klog.Infof("Finished reconciling ECSMService %s", key)
	return nil
}

// calculateStatus 是一个辅助函数，用于将现实世界的对象列表，聚合成 Status 结构
func (c *ECSMServiceController) calculateStatus(containers []clientset.ContainerInfo) ecsmv1.ECSMServiceStatus {
	var readyReplicas int32 = 0
	for _, c := range containers {
		if c.Status == "running" { // 假设 "running" 就是 "ready"
			readyReplicas++
		}
	}

	return ecsmv1.ECSMServiceStatus{
		Replicas:      int32(len(containers)),
		ReadyReplicas: readyReplicas,
		// TODO: 在这里填充 Conditions
	}
}
