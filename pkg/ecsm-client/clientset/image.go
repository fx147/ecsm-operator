package clientset

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

	ListAll(ctx context.Context, opts ImageListOptions) ([]ImageListItem, error)

	// GetDetails 根据镜像ID获取其详细信息。
	GetDetails(ctx context.Context, registryID, imageID string) (*ImageDetails, error)

	// GetDetailsByRef 是一个高级辅助函数，它封装了 "通过 ref 查找并获取详情" 的常用逻辑。
	GetDetailsByRef(ctx context.Context, registryID string, ref string) (*ImageDetails, error)

	// GetConfig 根据镜像ref获取其配置信息。
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

func (c *imageClient) ListAll(ctx context.Context, opts ImageListOptions) ([]ImageListItem, error) {
	var allItems []ImageListItem
	opts.PageNum = 1
	if opts.PageSize == 0 {
		opts.PageSize = 100
	}

	for {
		list, err := c.List(ctx, opts)
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

// GetDetailsByRef 实现了 ImageInterface 的同名方法。
func (c *imageClient) GetDetailsByRef(ctx context.Context, registryID, ref string) (*ImageDetails, error) {
	// 1. 解析 ref 字符串，获取 name, tag, os
	name, tag, os := parseRef(ref)
	if name == "" || tag == "" {
		return nil, fmt.Errorf("invalid image ref: '%s', expected format name@tag[#os]", ref)
	}

	// 2. 调用 ListAll 来获取所有可能的候选镜像
	// 我们只按 name 过滤，因为 tag 和 os 的匹配需要在客户端完成
	allImages, err := c.ListAll(ctx, ImageListOptions{
		RegistryID: registryID,
		Name:       name,
	})
	if err != nil {
		return nil, err
	}

	// 3. --- 核心修复：在列表中精确查找匹配的镜像 ---
	var foundImage *ImageListItem
	for i, img := range allImages {
		// 首先匹配 name 和 tag
		if img.Name == name && img.Tag == tag {
			// 如果 ref 中指定了 os，则必须匹配 os
			// 如果 ref 中没有指定 os，则匹配第一个找到的 name@tag
			if os != "" {
				if img.OS == os {
					foundImage = &allImages[i]
					break
				}
			} else {
				// 没有指定 os，第一个匹配的就是目标
				foundImage = &allImages[i]
				break
			}
		}
	}

	if foundImage == nil {
		return nil, fmt.Errorf("image with ref '%s' not found in registry '%s'", ref, registryID)
	}

	// 4. 找到后，用它的 ID 去调用底层的、更可靠的 GetDetails 方法
	return c.GetDetails(ctx, registryID, foundImage.ID)
}

// parseRef 是一个简单的 ref 解析器 (可以放在这个文件或一个 util 文件中)
func parseRef(ref string) (name, tag, os string) {
	parts := strings.SplitN(ref, "#", 2)
	if len(parts) == 2 {
		os = parts[1]
	}

	nameAndTag := parts[0]
	parts = strings.SplitN(nameAndTag, "@", 2)
	if len(parts) == 2 {
		name = parts[0]
		tag = parts[1]
	} else {
		name = parts[0] // 如果没有@tag，整个就是 name
	}
	return
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

func (i *ImageListItem) Ref() string {
	return fmt.Sprintf("%s@%s#%s", i.Name, i.Tag, i.OS)
}
