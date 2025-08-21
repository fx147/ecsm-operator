// file: pkg/registry/service.go

package registry

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	ecsmv1 "github.com/fx147/ecsm-operator/pkg/apis/ecsm/v1"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

var (
	_servicesBucketKey = []byte("ecsmservices")
)

func (r *Registry) CreateService(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error) {
	setServiceDefaults(service)
	if errs := validateService(service); len(errs) > 0 {
		return nil, errors.NewInvalid(ecsmv1.SchemeGroupVersion.WithKind("ECSMService").GroupKind(), service.Name, errs)
	}

	key, err := cache.MetaNamespaceKeyFunc(service)
	if err != nil {
		return nil, err
	}

	err = r.db.Update(func(tx *bolt.Tx) error {
		// 获取元数据和业务数据 bucket
		metaBucket := tx.Bucket(_metadataBucketKey)
		b, err := tx.CreateBucketIfNotExists(_servicesBucketKey)
		if err != nil {
			return err
		}

		// 检查对象是否已存在
		if b.Get([]byte(key)) != nil {
			return errors.NewAlreadyExists(ecsmv1.SchemeGroupVersion.WithResource("ecsmservices").GroupResource(), service.Name)
		}

		// 获取并递增全局 RV
		newRV, err := getAndIncrementGlobalRV(metaBucket)
		if err != nil {
			return err
		}

		// 填充系统字段
		service.ResourceVersion = strconv.FormatUint(newRV, 10)
		service.UID = types.UID(uuid.New().String())
		service.CreationTimestamp = metav1.Time{Time: time.Now().UTC()}

		buf, err := json.Marshal(service)
		if err != nil {
			return err
		}

		return b.Put([]byte(key), buf)
	})

	if err != nil {
		return nil, err
	}

	// 事务成功后，发布事件
	r.publish(Event{
		Type:            Added,
		Key:             key,
		Object:          service,
		ResourceVersion: service.ResourceVersion,
	})

	return service, nil
}

func (r *Registry) UpdateService(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error) {
	oldRVStr := service.ResourceVersion
	if oldRVStr == "" {
		errs := field.ErrorList{
			field.Required(field.NewPath("metadata", "resourceVersion"), "resourceVersion must be specified for an update"),
		}
		return nil, errors.NewInvalid(ecsmv1.SchemeGroupVersion.WithKind("ECSMService").GroupKind(), service.Name, errs)
	}

	key, err := cache.MetaNamespaceKeyFunc(service)
	if err != nil {
		return nil, err
	}

	err = r.db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket(_metadataBucketKey)
		b := tx.Bucket(_servicesBucketKey)
		if b == nil {
			return errors.NewNotFound(ecsmv1.SchemeGroupVersion.WithResource("ecsmservices").GroupResource(), service.Name)
		}

		// Check: 读取当前对象并比较 RV
		currentBytes := b.Get([]byte(key))
		if currentBytes == nil {
			return errors.NewNotFound(ecsmv1.SchemeGroupVersion.WithResource("ecsmservices").GroupResource(), service.Name)
		}

		var currentService ecsmv1.ECSMService
		if err := json.Unmarshal(currentBytes, &currentService); err != nil {
			return err
		}

		if currentService.ResourceVersion != oldRVStr {
			return errors.NewConflict(ecsmv1.SchemeGroupVersion.WithResource("ecsmservices").GroupResource(), service.Name, fmt.Errorf("object has been modified; please apply your changes to the latest version and try again"))
		}

		// Act: 递增 RV 并写入新对象
		newRV, err := getAndIncrementGlobalRV(metaBucket)
		if err != nil {
			return err
		}

		service.ResourceVersion = strconv.FormatUint(newRV, 10)
		// 确保 UID 和创建时间戳不被修改
		service.UID = currentService.UID
		service.CreationTimestamp = currentService.CreationTimestamp

		buf, err := json.Marshal(service)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), buf)
	})

	if err != nil {
		return nil, err
	}

	// 发布事件
	r.publish(Event{
		Type:            Modified,
		Key:             key,
		Object:          service,
		ResourceVersion: service.ResourceVersion,
	})

	return service, nil
}

// ... (List, Get, Delete 等方法的实现也应遵循类似的事务模式) ...
// UpdateServiceStatus 是一个专门用于更新 Service Status 子资源的业务方法。
// 它的核心逻辑是：只用传入对象的 status 覆盖存储中的 status，而 spec 和 metadata 保持不变。
func (r *Registry) UpdateServiceStatus(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error) {
	key, err := cache.MetaNamespaceKeyFunc(service)
	if err != nil {
		return nil, err
	}

	var updatedService *ecsmv1.ECSMService

	err = r.db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket(_metadataBucketKey)
		b := tx.Bucket(_servicesBucketKey)
		if b == nil {
			return errors.NewNotFound(ecsmv1.Resource("ecsmservices"), service.Name)
		}

		// 1. Get current object from store
		currentBytes := b.Get([]byte(key))
		if currentBytes == nil {
			return errors.NewNotFound(ecsmv1.Resource("ecsmservices"), service.Name)
		}

		var currentService ecsmv1.ECSMService
		if err := json.Unmarshal(currentBytes, &currentService); err != nil {
			return err
		}

		// 2. Prepare the object for update: copy spec and metadata from the stored object,
		//    and copy status from the incoming object.
		updatedService = currentService.DeepCopy() // Start with a deep copy of the current state
		updatedService.Status = service.Status     // Overwrite the status part

		// 3. Increment RV and write back
		newRV, err := getAndIncrementGlobalRV(metaBucket)
		if err != nil {
			return err
		}
		updatedService.ResourceVersion = strconv.FormatUint(newRV, 10)

		buf, err := json.Marshal(updatedService)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), buf)
	})

	if err != nil {
		return nil, err
	}

	// Publish the MODIFIED event with the fully updated object
	r.publish(Event{
		Type:            Modified,
		Key:             key,
		Object:          updatedService,
		ResourceVersion: updatedService.ResourceVersion,
	})

	return updatedService, nil
}

// GetService 是一个类型安全的方法，用于从 bbolt 中获取单个 ECSMService。
func (r *Registry) GetService(ctx context.Context, namespace, name string) (*ecsmv1.ECSMService, error) {
	key := namespace + "/" + name
	var service ecsmv1.ECSMService

	// 使用只读事务 (db.View) 进行读取，以获得更好的并发性能
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(_servicesBucketKey)
		if b == nil {
			return errors.NewNotFound(ecsmv1.Resource("ecsmservices"), name)
		}

		val := b.Get([]byte(key))
		if val == nil {
			return errors.NewNotFound(ecsmv1.Resource("ecsmservices"), name)
		}

		return json.Unmarshal(val, &service)
	})

	if err != nil {
		return nil, err
	}
	return &service, nil
}

// ListAllServices 返回指定命名空间下的所有 ECSMService 对象和一个全局的 ResourceVersion。
// 这个方法将用于 Informer 的 resync 过程。
func (r *Registry) ListAllServices(ctx context.Context, namespace string) (*ecsmv1.ECSMServiceList, string, error) {
	serviceList := &ecsmv1.ECSMServiceList{
		Items: []ecsmv1.ECSMService{},
	}
	var resourceVersion string

	err := r.db.View(func(tx *bolt.Tx) error {
		// --- 在同一个只读事务中，获取数据和全局版本号，保证一致性 ---

		// 1. 获取业务数据
		b := tx.Bucket(_servicesBucketKey)
		// 如果 bucket 不存在，说明没有任何 service，直接返回空列表
		if b == nil {
			return nil
		}

		c := b.Cursor()
		prefix := []byte(namespace + "/")

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var service ecsmv1.ECSMService
			if err := json.Unmarshal(v, &service); err != nil {
				// 记录错误但继续，以增加健壮性
				klog.Errorf("Failed to unmarshal service object with key %s: %v", string(k), err)
				continue
			}
			serviceList.Items = append(serviceList.Items, service)
		}

		// 2. 获取全局 ResourceVersion
		metaBucket := tx.Bucket(_metadataBucketKey)
		rvBytes := metaBucket.Get(_globalResourceVersionKey)
		if rvBytes != nil {
			rvUint := binary.BigEndian.Uint64(rvBytes)
			resourceVersion = strconv.FormatUint(rvUint, 10)
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return serviceList, resourceVersion, nil
}

// DeleteService ... (实现与 Create/Update 类似, 在 Update 事务中)
func (r *Registry) DeleteService(ctx context.Context, namespace, name string) error {
	key := namespace + "/" + name
	var deletedService ecsmv1.ECSMService

	err := r.db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket(_metadataBucketKey)
		b := tx.Bucket(_servicesBucketKey)
		if b == nil {
			return nil
		} // Already deleted

		// 在删除前获取对象，以便在事件中传递它
		val := b.Get([]byte(key))
		if val == nil {
			return nil
		} // Already deleted
		json.Unmarshal(val, &deletedService)

		if err := b.Delete([]byte(key)); err != nil {
			return err
		}

		// 删除也应该递增全局版本号
		_, err := getAndIncrementGlobalRV(metaBucket)
		return err
	})

	if err != nil {
		return err
	}

	r.publish(Event{
		Type:            Deleted,
		Key:             key,
		Object:          &deletedService,
		ResourceVersion: deletedService.ResourceVersion, // 传递被删除前的最后版本
	})

	return nil
}
func setServiceDefaults(service *ecsmv1.ECSMService) {
	// 填充默认值
}

func validateService(service *ecsmv1.ECSMService) field.ErrorList {
	// 验证对象
	return nil
}
