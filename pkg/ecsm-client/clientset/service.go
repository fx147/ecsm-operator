package clientset

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/rest"
)

type ServiceGetter interface {
	Services() ServiceInterface
}

// ServiceInterface 提供了所有操作 Service 核心资源的方法。
type ServiceInterface interface {
	// --- 核心 CRUD 操作 ---

	// Create 创建一个新的服务。
	Create(ctx context.Context, service *CreateServiceRequest) (*ServiceCreateResponse, error)

	// Get 根据服务 ID 获取一个服务的详细信息。
	Get(ctx context.Context, serviceID string) (*ServiceGet, error)

	// List 列出所有服务，支持通过 Options 进行过滤。
	List(ctx context.Context, opts ListServiceOptions) (*ServiceList, error)

	// Update 修改一个已存在的服务。
	Update(ctx context.Context, serviceID string, service *UpdateServiceRequest) (*ServiceCreateResponse, error)

	// Delete 根据服务 ID 删除一个服务。
	Delete(ctx context.Context, serviceID string) (*ServiceDeleteResponse, error)

	// // --- 批量操作 ---

	// // DeleteByPath 根据资源模板路径批量删除服务。
	// DeleteByPath(ctx context.Context, path string) error

	// // ControlByLabel 根据标签批量控制服务的状态 (start/stop/restart)。
	// ControlByLabel(ctx context.Context, labels map[string]string, action string) error

	// // --- 特殊操作 (Actions) ---

	// // Redeploy 触发一次服务的重新部署。
	// Redeploy(ctx context.Context, serviceID string) error

	// // ValidateName 校验服务名称是否合法或可用。
	// ValidateName(ctx context.Context, name string) (*ValidationResult, error)

	// // --- 状态与统计 ---

	// // GetStatistics 获取服务的统计信息。
	// GetStatistics(ctx context.Context) (*ServiceStatistics, error)
}

type serviceClient struct {
	restClient rest.Interface
}

func newService(restClient rest.Interface) *serviceClient {
	return &serviceClient{restClient: restClient}
}

// Create 实现了 ServiceInterface 的 Create 方法
func (c *serviceClient) Create(ctx context.Context, service *CreateServiceRequest) (*ServiceCreateResponse, error) {
	result := &ServiceCreateResponse{}

	// 开始构建请求
	err := c.restClient.Post().
		Resource("service").
		Body(service).
		Do(ctx).
		Into(result)

	return result, err
}

func (c *serviceClient) Update(ctx context.Context, serviceID string, service *UpdateServiceRequest) (*ServiceCreateResponse, error) {
	// 业务逻辑：确保传入的 serviceID 与 body 中的 ID 一致
	if serviceID != service.ID {
		return nil, fmt.Errorf("serviceID in path (%s) does not match serviceID in body (%s)", serviceID, service.ID)
	}

	result := &ServiceCreateResponse{}

	// 开始构建请求
	err := c.restClient.Put().
		Resource("service").
		Body(service).
		Do(ctx).
		Into(result)

	return result, err
}

func (c *serviceClient) Delete(ctx context.Context, serviceID string) (*ServiceDeleteResponse, error) {
	result := &ServiceDeleteResponse{}

	// 构建请求
	err := c.restClient.Delete().
		Resource("service").
		Name(serviceID).
		Do(ctx).
		Into(result)

	return result, err
}

func (c *serviceClient) Get(ctx context.Context, serviceID string) (*ServiceGet, error) {
	result := &ServiceGet{}

	// 开始构建请求
	err := c.restClient.Get().
		Resource("service").
		Name(serviceID).
		Do(ctx).
		Into(result)

	return result, err
}

// List 实现了 ServiceInterface 的 List 方法。
func (c *serviceClient) List(ctx context.Context, opts ListServiceOptions) (*ServiceList, error) {
	result := &ServiceList{}

	// 开始构建请求
	req := c.restClient.Get().Resource("service")

	// 将 Options 结构体翻译成 URL Query 参数
	req.Param("pageNum", strconv.Itoa(opts.PageNum))
	req.Param("pageSize", strconv.Itoa(opts.PageSize))
	if opts.Name != "" {
		req.Param("name", opts.Name)
	}
	if opts.ImageID != "" {
		// 注意：我们在 Go 结构体中叫 ImageID，但 API 参数是 id
		req.Param("id", opts.ImageID)
	}
	if opts.NodeID != "" {
		req.Param("nodeId", opts.NodeID)
	}
	if opts.Label != "" {
		req.Param("label", opts.Label)
	}

	// 执行请求并解码结果
	err := req.Do(ctx).Into(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ValidationResult 封装了名称校验的结果。
type ValidationResult struct {
	// IsValid bool
	// Reason  string
}

// ServiceStatistics 封装了服务的统计信息。
type ServiceStatistics struct {
	// ... (e.g., Total, Running, Deploying) ...
}
