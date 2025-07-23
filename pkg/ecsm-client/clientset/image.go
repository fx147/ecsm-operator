package clientset

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/rest"
)

// ImageGetter 提供了获取 Image 客户端的方法。
type ImageGetter interface {
	Images() ImageInterface
}

// ImageInterface 提供了所有查询 Image 资源的方法。
type ImageInterface interface {
	// List 列出镜像仓库中的所有镜像，支持通过 Options 进行过滤。
	List(ctx context.Context, opts ImageListOptions) (*ImageList, error)

	// GetDetails 根据镜像ID获取其详细信息。
	GetDetails(ctx context.Context, registryID, imageID string) (*ImageDetails, error)

	// GetConfig 根据镜像ID获取其配置信息。
	GetConfig(ctx context.Context, ref string) (*EcsImageConfig, error)

	// GetStatistics 获取镜像的统计信息（例如，总数）。
	GetStatistics(ctx context.Context) (*ImageStatistics, error)

	// GetRepositoryInfo 获取所有镜像仓库的信息和统计数据。
	// 支持通过 Options 进行过滤。
	GetRepositoryInfo(ctx context.Context, opts RepositoryInfoOptions) ([]RepositoryInfo, error)
}

type imageClient struct {
	restClient *rest.RESTClient
}

func newImages(restClient *rest.RESTClient) *imageClient {
	return &imageClient{restClient: restClient}
}

// List 实现了 ImageInterface 的同名方法。
func (c *imageClient) List(ctx context.Context, opts ImageListOptions) (*ImageList, error) {
	result := &ImageList{}

	req := c.restClient.Get().
		Resource("image")

	// 添加查询参数
	req.Param("registryId", opts.RegistryID) // 必填
	req.Param("pageNum", strconv.Itoa(opts.PageNum))
	req.Param("pageSize", strconv.Itoa(opts.PageSize))
	if opts.Name != "" {
		req.Param("name", opts.Name)
	}
	if opts.OS != "" {
		req.Param("os", opts.OS)
	}
	if opts.Author != "" {
		req.Param("author", opts.Author)
	}

	err := req.Do(ctx).Into(result)

	return result, err
}

func (c *imageClient) GetStatistics(ctx context.Context) (*ImageStatistics, error) {
	result := &ImageStatistics{}

	err := c.restClient.Get().
		Resource("image/summary").
		Do(ctx).
		Into(result)

	return result, err
}

// GetConfig 实现了 ImageInterface 的同名方法。
// 它会处理 API 响应中额外的 "config" 嵌套层，并直接返回 *EcsImageConfig。
func (c *imageClient) GetConfig(ctx context.Context, ref string) (*EcsImageConfig, error) {
	// 1. 定义一个临时的、匿名的结构体，它精确匹配 API 响应 data 字段的结构。
	//    这个结构体只在这个方法内部有效。
	responseWrapper := struct {
		Config *EcsImageConfig `json:"config"`
	}{} // 注意最后的 {} 创建了一个实例

	// 2. 构建请求
	req := c.restClient.Get().
		Resource("image/config").
		Param("ref", ref)

	// 3. 执行请求，并将 data 解码到我们临时的包装器中
	err := req.Do(ctx).Into(&responseWrapper)
	if err != nil {
		return nil, err
	}

	// 4. 检查内部的 Config 对象是否为 nil，避免 panic
	if responseWrapper.Config == nil {
		return nil, fmt.Errorf("API response for image config was successful, but the 'config' field is null")
	}

	// 5. 将解开包装的、真正的 *EcsImageConfig 对象返回给调用者
	return responseWrapper.Config, nil
}

// GetDetails 实现了 ImageInterface 的同名方法。
func (c *imageClient) GetDetails(ctx context.Context, registryID, imageID string) (*ImageDetails, error) {
	result := &ImageDetails{}

	err := c.restClient.Get().
		Resource("registry").
		Name(registryID).
		Subresource("image"). // 使用 Subresource 让语义更清晰
		Name(imageID).
		Do(ctx).
		Into(result)

	return result, err
}

// GetRepositoryInfo 实现了 ImageInterface 的同名方法。
func (c *imageClient) GetRepositoryInfo(ctx context.Context, opts RepositoryInfoOptions) ([]RepositoryInfo, error) {
	var result []RepositoryInfo

	req := c.restClient.Get().
		Resource("image/count")

	// 添加可选的查询参数
	if opts.Name != "" {
		req.Param("name", opts.Name)
	}
	if opts.OS != "" {
		req.Param("os", opts.OS)
	}
	if opts.Author != "" {
		req.Param("author", opts.Author)
	}

	err := req.Do(ctx).Into(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
