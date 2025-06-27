package registry

import (
	"context"

	ecsmv1 "github.com/fx147/ecsm-operator/pkg/apis/ecsm/v1"
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

// func (r *Registry) CreateService(ctx context.Context, service *ecsmv1.ECSMService) (*ecsmv1.ECSMService, error) {
// 	// 业务逻辑1 设置默认值
// 	setServiceDefaults(service)

// 	// 业务逻辑2 验证对象
// 	if errs := validateService(service); len(errs) > 0 {
// 		// 一般会返回一个包含所有错误的聚合错误
// 		// 这里为了简化，只返回第一个
// 		return nil, errors.NewInvalid(ecsmv1.Kind("ECSMService"), service.Name, errs)
// 	}
// }
