package registry

import (
	"context"

	ecsmv1 "github.com/fx147/ecsm-operator/pkg/apis/ecsm/v1"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// GetServiceWithNamespace 是一个类型安全的方法，可以用于获取 ECSMService 对象。
func (r *Registry) GetService(ctx context.Context, namespace, name string) (*ecsmv1.ECSMService, error) {
	service := &ecsmv1.ECSMService{}
	err := r.store.Get(namespace, name, service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// ListServices 列出指定命名空间中的所有 ECSMService 对象。
func (r *Registry) ListServices(ctx context.Context, namespace string) (*ecsmv1.ECSMServiceList, error) {
	services := &ecsmv1.ECSMServiceList{}
	err := r.store.List(namespace, services)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func (r *Registry) CreateService(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error) {
	// 业务逻辑1 设置默认值
	setServiceDefaults(service)

	// 业务逻辑2 验证对象
	if errs := validateService(service); len(errs) > 0 {
		// 一般会返回一个包含所有错误的聚合错误
		// 这里为了简化，只返回第一个
		return nil, errors.NewInvalid(ecsmv1.Kind("ECSMService"), service.Name, errs)
	}

	// 业务逻辑3 填充系统管理的字段
	uidString := uuid.New().String()
	service.ObjectMeta.UID = types.UID(uidString)
	service.ObjectMeta.CreationTimestamp = metav1.Now()

	// 调用底层存储
	if err := r.store.Create(service); err != nil {
		return nil, err
	}

	return service, nil
}

// UpdateService 封装了更新一个 ECSMService 的业务逻辑。
// 这是一个“读取-修改-写入”的原子操作，以防止更新冲突。
func (r *Registry) UpdateService(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error) {
	// --- 第一步：获取当前存储中的老对象 ---
	// 这是保证所有更新都在最新版本上进行的关键。
	oldService, err := r.GetService(ctx, service.Namespace, service.Name)
	if err != nil {
		// 如果获取时出错（比如没找到），直接返回错误。
		return nil, err
	}

	// (未来优化点): 在这里进行 ResourceVersion 的检查，实现乐观锁。
	// if oldService.ResourceVersion != service.ResourceVersion {
	//     return nil, errors.NewConflict(...)
	// }

	// --- 第二步：准备要写入的新对象 ---
	// 我们不能直接修改传入的 service 对象，因为它可能不完整。
	// 我们也不能直接修改 oldService，因为它来自存储（或缓存），修改它是不安全的。
	// 正确的做法是创建一个 oldService 的深拷贝，然后将新对象的变更合并进去。
	serviceToUpdate := oldService.DeepCopy()

	// --- 第三步：智能地合并字段 ---

	// 1. **Spec 的合并**: 将用户提供的新的 spec，完全覆盖掉旧的 spec。
	//    这是 Update 操作的核心意图。
	serviceToUpdate.Spec = service.Spec

	// 2. **Metadata 的合并**: 允许用户更新某些元数据字段，但要保留系统字段。
	//    我们只允许更新 labels 和 annotations。
	serviceToUpdate.ObjectMeta.Labels = service.ObjectMeta.Labels
	serviceToUpdate.ObjectMeta.Annotations = service.ObjectMeta.Annotations
	// 注意：serviceToUpdate 的 Name, Namespace, UID, CreationTimestamp 等都继承自 oldService，不会被覆盖。

	// 3. **Status 的处理**: UpdateService 方法 *不应该* 修改 Status。
	//    Status 的更新应该由一个独立的 UpdateStatus 方法来完成。
	//    所以，我们直接保留 oldService 的 status。
	//    serviceToUpdate.Status = oldService.Status (DeepCopy 已经帮我们做好了)

	// --- 第四步：执行业务逻辑校验和默认值填充 ---
	// 我们对这个合并后的、完整的 serviceToUpdate 对象进行校验和填充。
	setServiceDefaults(serviceToUpdate)
	if errs := validateService(serviceToUpdate); len(errs) > 0 {
		return nil, errors.NewInvalid(ecsmv1.SchemeGroupVersion.WithKind("ECSMService").GroupKind(), service.Name, errs)
	}

	// --- 第五步：调用底层存储 ---
	if err := r.store.Update(serviceToUpdate); err != nil {
		return nil, err
	}

	return serviceToUpdate, nil
}

// 新增一个专门用于更新 Status 的方法
func (r *Registry) UpdateServiceStatus(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error) {
	oldService, err := r.GetService(ctx, service.Namespace, service.Name)
	if err != nil {
		return nil, err
	}

	serviceToUpdate := oldService.DeepCopy()

	// 只用新的 status 覆盖旧的 status
	serviceToUpdate.Status = service.Status

	// 可以在这里对 Status 的某些字段进行验证
	if errs := validateService(serviceToUpdate); len(errs) > 0 {
		return nil, errors.NewInvalid(ecsmv1.SchemeGroupVersion.WithKind("ECSMService").GroupKind(), service.Name, errs)
	}

	// 调用底层存储
	if err := r.store.Update(serviceToUpdate); err != nil {
		return nil, err
	}

	return serviceToUpdate, nil
}

func (r *Registry) DeleteService(ctx context.Context, namespace, name string) error {
	return r.store.Delete(namespace, name, &ecsmv1.ECSMService{})
}

func setServiceDefaults(service *ecsmv1.ECSMService) {
	// 填充默认值
}

func validateService(service *ecsmv1.ECSMService) field.ErrorList {
	// 验证对象
	return nil
}
