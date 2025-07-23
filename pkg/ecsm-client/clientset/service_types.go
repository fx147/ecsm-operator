// file: pkg/ecsm_client/clientset/service_types.go

package clientset

// --- Create Request Structures ---

// CreateServiceRequest 完整地定义了创建一个新服务时，ECSM API 所需的 payload。
type CreateServiceRequest struct {
	Name    string    `json:"name"`
	Image   ImageSpec `json:"image"`
	Node    NodeSpec  `json:"node"`
	Factor  *int      `json:"factor,omitempty"`
	Policy  string    `json:"policy,omitempty"` // "dynamic" or "static"
	Prepull *bool     `json:"prepull,omitempty"`
}

type ImageSpec struct {
	Ref         string          `json:"ref"`
	Action      string          `json:"action"` // "load" or "run"
	Config      *EcsImageConfig `json:"config"` // 假设我们只关心 EcsImageConfig
	VSOA        *ImageVSOA      `json:"vsoa,omitempty"`
	PullPolicy  string          `json:"pullPolicy,omitempty"`
	AutoUpgrade string          `json:"autoUpgrade,omitempty"`
}

type NodeSpec struct {
	Names []string `json:"names"`
}

type ImageVSOA struct {
	Password          string `json:"password,omitempty"`
	Port              *int   `json:"port,omitempty"`
	HealthPath        string `json:"healthPath,omitempty"`
	HealthTimeout     *int   `json:"healthTimeout,omitempty"`
	HealthRetries     *int   `json:"healthRetries,omitempty"`
	HealthStartPeriod *int   `json:"healthStartPeriod,omitempty"`
	HealthInterval    *int   `json:"healthInterval,omitempty"`
}

// --- Create Response Structures ---

// ServiceInfo 代表从 API 的 Create, Get 或 List 调用中返回的单个服务的核心信息。
// 注意：这个结构是根据 Create 的响应来定义的。Get 和 List 的响应可能会更丰富，
type ServiceCreateResponse struct {
	ID         string   `json:"id"`
	Containers []string `json:"containers"`
}

type ServiceDeleteResponse struct {
	ID string `json:"transactionId"`
}

// ServiceGet mimics the response from the GET /service/:id endpoint.
// ServiceGet 精确匹配 GET /service/:id API 的成功响应 data。
type ServiceGet struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Status               string            `json:"status"`
	ContainerStatusGroup []string          `json:"containerStatusGroup"`
	Healthy              bool              `json:"healthy"`
	Factor               int               `json:"factor"`
	Policy               string            `json:"policy"`
	InstanceOnline       int               `json:"instanceOnline"`
	InstanceActive       int               `json:"instanceActive"`
	CreatedTime          string            `json:"createdTime"`
	UpdatedTime          string            `json:"updatedTime"`
	Image                *ImageSpec        `json:"image"`          // <-- 复用共享类型
	Node                 *NodeSpec         `json:"node,omitempty"` // <-- 复用共享类型
	NodeList             []ServiceNodeInfo `json:"nodeList"`
}

// --- List Options and Response Structures ---
// ListServiceOptions 封装了所有可以用于 List 服务的查询参数。
type ListServicesOptions struct {
	PageNum  int    `json:"pageNum"`  // 必填
	PageSize int    `json:"pageSize"` // 必填
	Name     string `json:"name,omitempty"`
	// 注意：API 文档中的 'id' 字段名可能会引起混淆，因为它指的是镜像ID，
	// 我们在结构体中用更明确的名字 ImageID。
	ImageID string `json:"imageId,omitempty"`
	NodeID  string `json:"nodeId,omitempty"`
	Label   string `json:"label,omitempty"`
}

// ServiceList 是 List 方法的返回值，精确匹配 API 响应中的 data 字段。
type ServiceList struct {
	Total    int                `json:"total"`
	PageNum  int                `json:"pageNum"`
	PageSize int                `json:"pageSize"`
	Items    []ProvisionListRow `json:"list"` // 字段名是 "list"
}

// ProvisionListRow 代表服务列表中的单行数据。
type ProvisionListRow struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Status               string            `json:"status"`
	UpdatedTime          string            `json:"updatedTime"`
	CreatedTime          string            `json:"createdTime"`
	ImageList            []ImageListEntry  `json:"imageList"`
	NodeList             []ServiceNodeInfo `json:"nodeList"`
	ContainerStatusGroup []string          `json:"containerStatusGroup"`
	Factor               int               `json:"factor"`
	Policy               string            `json:"policy"`
	ErrorInstances       []ErrorInstance   `json:"errorInstance"`
	InstanceOnline       int               `json:"instanceOnline"`
	DefaultLabels        []string          `json:"defaultLabels"`
	PathLabel            string            `json:"pathLabel"`
}

// ImageListEntry 是服务列表中内嵌的镜像信息。
type ImageListEntry struct {
	Name string `json:"name"`
	OS   string `json:"os"`
	Tag  string `json:"tag"`
}

// NodeListEntry 是服务列表中内嵌的节点信息。
type ServiceNodeInfo struct {
	NodeID   string `json:"nodeId"`
	NodeName string `json:"nodeName"`
	Address  string `json:"address"`
}

// ErrorInstance 描述了一个部署失败的实例。
type ErrorInstance struct {
	ContainerID string `json:"containerId"`
	NodeID      string `json:"nodeId"`
	NodeName    string `json:"nodeName"`
	Status      bool   `json:"status"` // 文档写的是string，但含义是bool，我们先用bool
	Message     string `json:"message"`
}

// --- Update Request Structures ---

// UpdateServiceRequest 定义了更新一个服务时，ECSM API 所需的 payload。
// 它与 CreateServiceRequest 非常相似，但包含了服务ID。
type UpdateServiceRequest struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Image  ImageSpec `json:"image"`
	Node   NodeSpec  `json:"node"`
	Factor *int      `json:"factor,omitempty"`
	Policy string    `json:"policy,omitempty"` // "dynamic" or "static"

	// 注意：Update 的 payload 中似乎没有 prepull 字段，所以我们不在这里包含它。
}
