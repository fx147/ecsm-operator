package clientset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/rest"
)

type NodeGetter interface {
	Nodes() NodeInterface
}

type NodeInterface interface {
	// --- 核心 CRUD 操作 ---
	// Register 注册一个新的节点。
	Register(ctx context.Context, req *NodeRegisterRequest) error

	// ValidateName 校验节点名称是否可用。
	ValidateName(ctx context.Context, opts NodeValidateNameOptions) (*ValidationResult, error)

	// ValidateAddress 校验节点地址是否可用。
	ValidateAddress(ctx context.Context, opts NodeValidateAddressOptions) (*ValidationResult, error)

	// Update 修改一个已存在的节点, 成功时不返回节点信息，只返回 error
	Update(ctx context.Context, nodeID string, req *NodeUpdateRequest) error

	// RefreshNodeTypes 触发一个后台任务，更新所有节点的类型信息。
	// 这是一个异步触发器，成功时只表示任务已提交。
	RefreshNodeTypes(ctx context.Context) error

	// CheckNodeTypeUpdates 查询所有节点的类型更新状态。
	// 它返回一个列表，其中包含了正在变更或已变更类型的节点信息。
	CheckNodeTypeUpdates(ctx context.Context) ([]NodeTypeUpdateInfo, error)

	List(ctx context.Context, opts NodeListOptions) (*NodeList, error)

	ListAll(ctx context.Context, opts NodeListOptions) ([]NodeInfo, error)

	GetByID(ctx context.Context, nodeID string) (*NodeDetailsByID, error)

	GetByName(ctx context.Context, nodeName string) (*NodeDetailsByName, error) // 返回 *NodeDetailsByName

	GetNodeView(ctx context.Context, nodeID string) (*NodeView, error)

	GetNodeMetrics(ctx context.Context, opts NodeMetricsOptions) ([]NodeMetrics, error)

	// ListStatus 根据一组节点 ID，批量获取它们的实时运行时状态。
	ListStatus(ctx context.Context, nodeIDs []string) ([]NodeStatus, error)

	// Delete 批量删除一个或多个节点。
	// 如果删除操作因为节点被占用而部分或全部失败，
	// 它会返回一个非空的冲突列表和一个 nil 错误。
	// 只有在发生网络错误或 API 返回非 200 状态时，才会返回非 nil 的 error。
	Delete(ctx context.Context, nodeIDs []string) ([]NodeDeleteConflict, error)
}

type nodeClient struct {
	restClient rest.Interface
}

func newNodes(c rest.Interface) *nodeClient {
	return &nodeClient{restClient: c}
}

func (c *nodeClient) Register(ctx context.Context, req *NodeRegisterRequest) error {
	// 我们不期望有任何结构化的 data 返回，所以 Into(nil) 是完美的。
	// Into(nil) 会处理 status!=200 的情况，如果成功，则直接返回 nil。
	err := c.restClient.Post().
		Resource("node").
		Body(req).
		Do(ctx).
		Into(nil)

	return err
}

// ValidateName 实现了 NodeInterface 的同名方法。
func (c *nodeClient) ValidateName(ctx context.Context, opts NodeValidateNameOptions) (*ValidationResult, error) {
	// 准备一个用于接收解码后 data (一个布尔值) 的容器
	var nameExists bool

	// 开始构建请求
	req := c.restClient.Get().
		Resource("node/name/check")

	// 添加查询参数
	req.Param("name", opts.Name)
	if opts.ExcludeID != "" {
		req.Param("id", opts.ExcludeID)
	}

	err := req.Do(ctx).Into(&nameExists)
	if err != nil {
		return nil, err
	}

	// 将 API 返回的 "exists" (存在) 逻辑，转换为我们更通用的 "IsValid" (有效) 逻辑
	// 如果 nameExists 为 true，说明名称已存在，即名称无效 (IsValid = false)
	result := &ValidationResult{
		IsValid: !nameExists,
	}

	if nameExists {
		result.Message = fmt.Sprintf("node name '%s' already exists", opts.Name)
	}

	return result, nil
}

func (c *nodeClient) ValidateAddress(ctx context.Context, opts NodeValidateAddressOptions) (*ValidationResult, error) {
	// 准备一个用于接收解码后 data (一个布尔值) 的容器
	var addressExists bool

	// 开始构造请求
	req := c.restClient.Get().
		Resource("node/address/check")

	// 添加查询参数
	req.Param("address", opts.Address)
	if opts.ExcludeID != "" {
		req.Param("id", opts.ExcludeID)
	}
	if opts.TLS != nil {
		req.Param("tls", strconv.FormatBool(*opts.TLS))
	}

	err := req.Do(ctx).Into(&addressExists)
	if err != nil {
		return nil, err
	}

	// 将 API 返回的 "exists" (存在) 逻辑，转换为我们更通用的 "IsValid" (有效) 逻辑
	// 如果 addressExists 为 true，说明地址已存在，即地址无效 (IsValid = false)
	result := &ValidationResult{
		IsValid: !addressExists,
	}

	if addressExists {
		result.Message = fmt.Sprintf("node address '%s' already exists", opts.Address)
	}

	return result, nil
}

// Update 实现了 NodeInterface 的同名方法。
func (c *nodeClient) Update(ctx context.Context, nodeID string, req *NodeUpdateRequest) error {
	// 最佳实践：进行一次健全性检查
	if nodeID != req.ID {
		return fmt.Errorf("nodeID in path (%s) does not match ID in request body (%s)", nodeID, req.ID)
	}

	// 我们不期望有任何结构化的 data 返回，所以 Into(nil) 是完美的。
	err := c.restClient.Put().
		Resource("node").
		Body(req).
		Do(ctx).
		Into(nil)

	return err
}

// RefreshNodeTypes 实现了 NodeInterface 的同名方法。
func (c *nodeClient) RefreshNodeTypes(ctx context.Context) error {
	// 这个请求没有 body，所以 Body(nil)
	// 我们不期望有任何结构化的 data 返回，所以 Into(nil)
	err := c.restClient.Put().
		Resource("node/type").
		Body(nil). // 无请求体
		Do(ctx).
		Into(nil) // 只关心成功或失败

	return err
}

// CheckNodeTypeUpdates 实现了 NodeInterface 的同名方法。
func (c *nodeClient) CheckNodeTypeUpdates(ctx context.Context) ([]NodeTypeUpdateInfo, error) {
	// 准备一个用于接收解码后 data (一个对象数组) 的容器
	var result []NodeTypeUpdateInfo

	// 这个请求没有 body 和 query 参数
	err := c.restClient.Get().
		Resource("node/type/check").
		Do(ctx).
		Into(&result)

	return result, err
}

// List 实现了 NodeInterface 的同名方法。
func (c *nodeClient) List(ctx context.Context, opts NodeListOptions) (*NodeList, error) {
	result := &NodeList{}
	req := c.restClient.Get().Resource("node")

	req.Param("pageNum", strconv.Itoa(opts.PageNum))
	req.Param("pageSize", strconv.Itoa(opts.PageSize))
	if opts.Name != "" {
		req.Param("name", opts.Name)
	}
	if opts.BasicInfo {
		req.Param("basicInfo", "true")
	}

	err := req.Do(ctx).Into(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *nodeClient) GetByID(ctx context.Context, nodeID string) (*NodeDetailsByID, error) {
	result := &NodeDetailsByID{}

	err := c.restClient.Get().
		Resource("node").
		Name(nodeID).
		Do(ctx).
		Into(result)

	return result, err
}

func (c *nodeClient) GetByName(ctx context.Context, nodeName string) (*NodeDetailsByName, error) {
	result := &NodeDetailsByName{}

	err := c.restClient.Get().
		Resource("node/name").
		Name(nodeName).
		Do(ctx).
		Into(result)

	return result, err
}

// ListStatus 实现了 NodeInterface 的同名方法。
func (c *nodeClient) ListStatus(ctx context.Context, nodeIDs []string) ([]NodeStatus, error) {
	// 准备一个用于接收解码后 data 字段的容器
	result := &NodeStatusResponse{}

	req := c.restClient.Get().
		Resource("node/status")

	// 将 nodeIDs 切片编码为多个 ids[]=<id> 的查询参数
	for _, id := range nodeIDs {
		req.Param("ids[]", id)
	}

	err := req.Do(ctx).Into(result)
	if err != nil {
		return nil, err
	}

	// 返回 Nodes 列表，而不是整个响应结构体
	return result.Nodes, nil
}

// Delete 实现了 NodeInterface 的同名方法。
func (c *nodeClient) Delete(ctx context.Context, nodeIDs []string) ([]NodeDeleteConflict, error) {
	// 构造请求体
	reqBody := &NodeDeleteRequest{
		IDs: nodeIDs,
	}

	// 1. 执行请求并获取原始的响应体 []byte
	respBody, err := c.restClient.Delete().
		Resource("node").
		Body(reqBody).
		Do(ctx).
		Raw()
	if err != nil {
		return nil, err
	}

	// 2. 将响应体解码到我们导出的 rest.Response 结构体中
	var apiResp rest.Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode generic response: %w", err)
	}

	// 3. 检查 API 级别的错误
	if apiResp.Status != 200 {
		// --- 核心修复 2 ---
		// 构造并返回一个实现了 error 接口的 *rest.Aerror
		return nil, &rest.Aerror{
			Status:      apiResp.Status,
			Message:     apiResp.Message,
			FieldErrors: apiResp.FieldErrors,
		}
	}

	// 4. 探测 data 字段的类型
	trimmedData := bytes.TrimSpace(apiResp.Data)
	if len(trimmedData) == 0 || string(trimmedData) == "null" {
		return nil, fmt.Errorf("delete response data is empty or null, which is unexpected")
	}

	if bytes.HasPrefix(trimmedData, []byte{'['}) {
		// 这是一个冲突列表
		var conflicts []NodeDeleteConflict
		if err := json.Unmarshal(trimmedData, &conflicts); err != nil {
			return nil, fmt.Errorf("failed to unmarshal delete conflicts: %w", err)
		}
		return conflicts, nil
	}

	if bytes.HasPrefix(trimmedData, []byte{'"'}) {
		// 这是一个字符串
		var successMsg string
		if err := json.Unmarshal(trimmedData, &successMsg); err == nil && successMsg == "success" {
			// 完全成功，返回一个空的冲突列表和 nil 错误
			return nil, nil
		}
	}

	return nil, fmt.Errorf("unexpected data format in delete response: %s", string(trimmedData))
}

// ListAll 实现了 NodeInterface 的同名方法。
func (c *nodeClient) ListAll(ctx context.Context, opts NodeListOptions) ([]NodeInfo, error) {
	var allNodes []NodeInfo
	// 确保 PageNum 从 1 开始
	opts.PageNum = 1

	// 如果用户没有指定 PageSize，我们用一个较大的默认值来提高效率
	if opts.PageSize == 0 {
		opts.PageSize = 100
	}

	for {
		// 调用同一个客户端的 List 方法获取一页数据
		list, err := c.List(ctx, opts)
		if err != nil {
			return nil, err
		}

		// 如果当前页没有任何数据，说明已经结束
		if len(list.Items) == 0 {
			break
		}

		allNodes = append(allNodes, list.Items...)

		// 检查是否已获取所有
		if len(allNodes) >= list.Total {
			break
		}

		// 准备获取下一页
		opts.PageNum++
	}
	return allNodes, nil
}

func (c *nodeClient) GetNodeView(ctx context.Context, nodeID string) (*NodeView, error) {
	result := &NodeView{}
	err := c.restClient.Get().
		Resource("overview/platform/node-view").
		Name(nodeID).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *nodeClient) GetNodeMetrics(ctx context.Context, opts NodeMetricsOptions) ([]NodeMetrics, error) {
	var result []NodeMetrics
	req := c.restClient.Get().
		Resource("overview/node").
		Param("nodeId", opts.NodeID).
		Param("instant", strconv.FormatBool(opts.Instant))
	// ... (add other optional params)
	err := req.Do(ctx).Into(&result)
	return result, err
}
