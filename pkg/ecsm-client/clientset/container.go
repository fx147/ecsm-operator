package clientset

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/rest"
)

type ContainerGetter interface {
	Containers() ContainerInterface
}

type ContainerInterface interface {
	GetByTaskID(ctx context.Context, taskId string) (*ContainerInfo, error)

	GetByName(ctx context.Context, serviceClient ServiceInterface, name string) (*ContainerInfo, error)

	// GetByTaskID 根据容器的 *任务ID* 获取其详细信息。
	GetHistory(ctx context.Context, opts ContainerHistoryOptions) (*ContainerHistoryList, error)

	ListByService(ctx context.Context, opts ListContainersByServiceOptions) (*ContainerList, error)

	ListAllByService(ctx context.Context, opts ListContainersByServiceOptions) ([]ContainerInfo, error)

	ListByNode(ctx context.Context, opts ListContainersByNodeOptions) (*ContainerList, error)

	ListAllByNode(ctx context.Context, opts ListContainersByNodeOptions) ([]ContainerInfo, error)

	SubmitControlActionByName(ctx context.Context, containerName string, action ContainerAction) (*Transaction, error)

	SubmitControlActionByService(ctx context.Context, serviceID string, action ContainerAction) (*Transaction, error)
}

type containerClient struct {
	restClient rest.Interface
}

func newContainers(restClient rest.Interface) *containerClient {
	return &containerClient{restClient: restClient}
}

func (c *containerClient) GetByTaskID(ctx context.Context, taskId string) (*ContainerInfo, error) {
	result := &ContainerInfo{}
	err := c.restClient.Get().
		Resource("container").
		Name(taskId).
		Do(ctx).
		Into(result)
	return result, err
}

// ListByService 实现了 ContainerInterface 的 ListByService 方法。
func (c *containerClient) ListByService(ctx context.Context, opts ListContainersByServiceOptions) (*ContainerList, error) {
	result := &ContainerList{}

	req := c.restClient.Get().Resource("container/service")

	// 添加查询参数
	req.Param("pageNum", strconv.Itoa(opts.PageNum))
	req.Param("pageSize", strconv.Itoa(opts.PageSize))
	if opts.Key != "" {
		req.Param("key", opts.Key)
	}

	// 特别处理 string 数组参数
	// ECSM API 期望的格式是 serviceIds[]=...&serviceIds[]=...
	// url.Values 的 Add 方法默认就能处理好这个
	for _, id := range opts.ServiceIDs {
		req.Param("serviceIds[]", id)
	}

	err := req.Do(ctx).Into(result)
	return result, err
}

func (c *containerClient) ListByNode(ctx context.Context, opts ListContainersByNodeOptions) (*ContainerList, error) {
	result := &ContainerList{}
	req := c.restClient.Get().Resource("container/node")

	// 添加查询参数
	req.Param("pageNum", strconv.Itoa(opts.PageNum))
	req.Param("pageSize", strconv.Itoa(opts.PageSize))
	if opts.Key != "" {
		req.Param("key", opts.Key)
	}

	for _, id := range opts.NodeIDs {
		req.Param("nodeIds[]", id)
	}

	err := req.Do(ctx).Into(result)
	return result, err
}

// SubmitControlActionByName 实现了 ContainerInterface 的同名方法。
func (c *containerClient) SubmitControlActionByName(ctx context.Context, containerName string, action ContainerAction) (*Transaction, error) {
	// 构造请求体
	reqBody := &ContainerControlByNameRequest{
		Name:   containerName, // 传入的参数是 name
		Action: action,
	}

	result := &Transaction{}

	err := c.restClient.Put().
		Resource("container").
		Body(reqBody).
		Do(ctx).
		Into(result) // 将返回的 data 解码到 Transaction 对象中

	return result, err
}

func (c *containerClient) SubmitControlActionByService(ctx context.Context, serviceID string, action ContainerAction) (*Transaction, error) {
	// 构造请求体
	reqBody := &ServiceControlContainerRequest{
		ID:     serviceID, // 传入的参数是 serviceId
		Action: action,
	}

	result := &Transaction{}

	err := c.restClient.Put().
		Resource("service/container").
		Body(reqBody).
		Do(ctx).
		Into(result) // 将返回的 data 解码到 Transaction 对象中

	return result, err
}

// GetHistory 实现了 ContainerInterface 的同名方法。
func (c *containerClient) GetHistory(ctx context.Context, opts ContainerHistoryOptions) (*ContainerHistoryList, error) {
	result := &ContainerHistoryList{}

	req := c.restClient.Get().
		// 注意 URL 路径是 "container/action/history"
		Resource("container/action/history")

	// 添加查询参数
	req.Param("pageNum", strconv.Itoa(opts.PageNum))
	req.Param("pageSize", strconv.Itoa(opts.PageSize))
	req.Param("id", opts.TaskID) // 将 TaskID 翻译回 'id' 参数

	err := req.Do(ctx).Into(result)
	return result, err
}

func (c *containerClient) ListAllByService(ctx context.Context, opts ListContainersByServiceOptions) ([]ContainerInfo, error) {
	var allItems []ContainerInfo
	opts.PageNum = 1
	if opts.PageSize == 0 {
		opts.PageSize = 100
	}

	for {
		list, err := c.ListByService(ctx, opts)
		if err != nil {
			return nil, err
		}

		if len(list.Items) == 0 {
			break
		}

		allItems = append(allItems, list.Items...)

		if len(allItems) >= list.Total {
			break
		}

		opts.PageNum++
	}
	return allItems, nil
}

func (c *containerClient) ListAllByNode(ctx context.Context, opts ListContainersByNodeOptions) ([]ContainerInfo, error) {
	var allItems []ContainerInfo
	opts.PageNum = 1
	if opts.PageSize == 0 {
		opts.PageSize = 100
	}

	for {
		list, err := c.ListByNode(ctx, opts)
		if err != nil {
			return nil, err
		}

		if len(list.Items) == 0 {
			break
		}

		allItems = append(allItems, list.Items...)

		if len(allItems) >= list.Total {
			break
		}

		opts.PageNum++
	}
	return allItems, nil
}

func (c *containerClient) GetByName(ctx context.Context, serviceClient ServiceInterface, name string) (*ContainerInfo, error) {
	// 1. 获取所有服务
	allServices, err := serviceClient.ListAll(ctx, ListServicesOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all services to find container: %w", err)
	}

	var allServiceIDs []string
	for _, svc := range allServices {
		allServiceIDs = append(allServiceIDs, svc.ID)
	}

	if len(allServiceIDs) == 0 {
		return nil, fmt.Errorf("no services found in the system")
	}

	// 2. 获取所有服务下的所有容器
	allContainers, err := c.ListAllByService(ctx, ListContainersByServiceOptions{ServiceIDs: allServiceIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to list all containers: %w", err)
	}

	// 3. 查找匹配的容器
	for i, container := range allContainers {
		if container.Name == name {
			return &allContainers[i], nil
		}
	}

	return nil, fmt.Errorf("container with name '%s' not found", name)
}
