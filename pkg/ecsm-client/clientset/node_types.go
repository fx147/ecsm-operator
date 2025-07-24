package clientset

// NodeRegisterRequest 定义了注册一个新节点时所需的 payload。
type NodeRegisterRequest struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Password string `json:"password"`
	// 使用指针以便 omitempty 能正确处理 false 值
	TLS *bool `json:"tls,omitempty"`
}

// --- Node Action-Specific Structures ---

// NodeValidateNameOptions 封装了校验节点名称时可以传入的参数。
type NodeValidateNameOptions struct {
	// Name 是要校验的节点名称。
	Name string
	// ExcludeID 是一个可选的节点ID，在校验时会排除这个ID对应的节点。
	// 这在“更新”一个节点时检查新名称是否与“其他”节点冲突时非常有用。
	ExcludeID string
}

// NodeValidateAddressOptions 封装了校验节点地址时可以传入的参数。
type NodeValidateAddressOptions struct {
	// Address 是要校验的节点地址。
	Address string
	// ExcludeID 是一个可选的节点ID，在校验时会排除这个ID对应的节点。
	ExcludeID string
	// TLS 指定校验时是否考虑 TLS 加密。
	TLS *bool
}

// ValidationResult 封装了一个验证操作的结果。
// 注意：这个结构体可以被多个 validate* 方法复用。
type ValidationResult struct {
	// IsValid 表示验证是否通过。
	// 对于 name/check，如果名称已存在，我们会把它转换为 IsValid=false。
	IsValid bool
	// Message 提供了验证失败的额外信息。
	Message string
}

// NodeUpdateRequest 定义了修改一个节点时所需的 payload。
type NodeUpdateRequest struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Name     string `json:"name"`
	Password string `json:"password"`
	TLS      bool   `json:"tls"` // 文档说是必填，所以直接用 bool
}

// NodeTypeUpdateInfo 描述了单个节点的类型更新状态。
type NodeTypeUpdateInfo struct {
	ID      string `json:"id"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	NewType string `json:"newType"`
}

// --- Node Get/List Structures ---

// NodeListOptions 封装了所有可以用于 List 节点的查询参数。
type NodeListOptions struct {
	PageNum   int
	PageSize  int
	Name      string
	BasicInfo bool
}

// NodeList 是 List 方法的返回值，精确匹配 API 响应的 data 字段。
type NodeList struct {
	Total    int        `json:"total"`
	PageNum  int        `json:"pageNum"`
	PageSize int        `json:"pageSize"`
	Items    []NodeInfo `json:"list"` // 注意：Items 的类型是 NodeInfo
}

// NodeInfo 代表节点列表中的单个节点运行时信息 (basicInfo=false 时)。
type NodeInfo struct {
	ID                   string  `json:"id"`
	Address              string  `json:"address"`
	Name                 string  `json:"name"`
	Password             string  `json:"password,omitempty"` // List 的真实响应中包含 password
	Status               string  `json:"status"`
	Type                 string  `json:"type"`
	TLS                  bool    `json:"tls"`
	ContainerTotal       int     `json:"containerTotal"`
	ContainerRunning     int     `json:"containerRunning"`
	ContainerEcsmTotal   int     `json:"containerEcsmTotal"`
	ContainerEcsmRunning int     `json:"containerEcsmRunning"`
	UpTime               float64 `json:"upTime"`
	CreatedTime          string  `json:"createdTime"`
	Arch                 string  `json:"arch"`
}

// NodeDetails 代表通过 Get /node/:id 获取到的节点详细配置信息。
type NodeDetailsByID struct {
	ID          string `json:"id"`
	Address     string `json:"address"`
	Name        string `json:"name"`
	Password    string `json:"password"`
	TLS         bool   `json:"tls"`
	Type        string `json:"type"`
	CreatedTime string `json:"createdTime"`
	Arch        string `json:"arch"`
	EcsdVersion string `json:"ecsdVersion"` // Get 详情时特有的字段
}

// NodeDetails 代表通过 Get /node/:name 获取到的节点详细配置信息。
type NodeDetailsByName struct {
	ID          string `json:"id"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Name        string `json:"name"`
	Password    string `json:"password"`
	TLS         int    `json:"tls"`
	Type        string `json:"type"`
	CreatedTime string `json:"createdTime"`
	Arch        string `json:"arch"`
}

// NodeStatus 描述了一个节点的实时运行时状态，精确匹配 GET /node/status API 的响应。
type NodeStatus struct {
	ID                   string        `json:"id"`
	Status               string        `json:"status"`
	MemoryTotal          int64         `json:"memoryTotal"`
	MemoryFree           int64         `json:"memoryFree"`
	DiskTotal            float64       `json:"diskTotal"`
	DiskFree             float64       `json:"diskFree"`
	CPUUsage             NodeCPUUsage  `json:"cpuUsage"`
	Uptime               float64       `json:"uptime"`
	ProcessCount         int           `json:"processCount"`
	ContainerTotal       int           `json:"containerTotal"`
	ContainerRunning     int           `json:"containerRunning"`
	ContainerEcsmTotal   int           `json:"containerEcsmTotal"`
	ContainerEcsmRunning int           `json:"containerEcsmRunning"`
	Net                  []NodeNetInfo `json:"net"`
	Time                 NodeTimeInfo  `json:"time"`
}

// NodeCPUUsage 描述了节点的 CPU 使用情况。
type NodeCPUUsage struct {
	Total       float64   `json:"total"`
	Cores       []float64 `json:"cores"`
	MeasureTime int       `json:"measureTime"` // 从响应示例中补充
}

// NodeNetInfo 描述了节点的网络接口情况。
type NodeNetInfo struct {
	NetworkName string  `json:"networkName"`
	UpNet       float64 `json:"upNet"`
	DownNet     float64 `json:"downNet"`
}

// NodeTimeInfo 描述了节点的时区和时间信息。
type NodeTimeInfo struct {
	Current      int64   `json:"current"`
	Uptime       float64 `json:"uptime"`
	Timezone     string  `json:"timezone"`
	TimezoneName string  `json:"timezoneName"`
	Date         string  `json:"date"` // 从响应示例中补充
}

// NodeStatusResponse 是 GET /node/status API 响应的 data 字段。
type NodeStatusResponse struct {
	Nodes []NodeStatus `json:"nodes"`
}

// NodeDeleteRequest 定义了批量删除节点时所需的 payload。
type NodeDeleteRequest struct {
	IDs []string `json:"ids"`
}

// NodeDeleteConflict 描述了一个因为被服务占用而无法删除的节点。
type NodeDeleteConflict struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	Serves []ConflictingService `json:"serves"`
}

// ConflictingService 描述了占用节点的具体服务。
type ConflictingService struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// --- NodeView Structures ---
type NodeView struct {
	ID       string              `json:"id"`
	Status   string              `json:"status"`
	Type     string              `json:"type"`
	Name     string              `json:"name"`
	Children []NodeViewContainer `json:"children"`
	// ... (其他静态字段如 ip, arch)
}
type NodeViewContainer struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	NodeID    string              `json:"node_id"`
	ServiceID string              `json:"pro_id"`
	Type      string              `json:"type"`
	Status    string              `json:"status"`
	Children  []NodeViewProvision `json:"children"`
}
type NodeViewProvision struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Health bool   `json:"health"`
}

// --- NodeMetrics Structures ---
type NodeMetrics struct {
	Timestamp    int64               `json:"timestamp"`
	Type         string              `json:"type"`
	CPU          MetricValue         `json:"cpu"`
	ROM          MetricValueWithSize `json:"rom"`
	RAM          MetricValueWithSize `json:"ram"`
	ProcessCount int                 `json:"processCount"`
	Processes    []ProcessMetrics    `json:"process"`
	UpNet        []NetMetrics        `json:"upNet"`
	DownNet      []NetMetrics        `json:"downNet"`
	Uptime       float64             `json:"upTime"`
	Running      int                 `json:"running"`
	Stop         int                 `json:"stop"`
}
type MetricValue struct {
	Percent string `json:"percent"`
}
type MetricValueWithSize struct {
	Percent string  `json:"percent"`
	Size    float64 `json:"size"`
}
type ProcessMetrics struct {
	Name string              `json:"name"`
	CPU  MetricValue         `json:"cpu"`
	RAM  MetricValueWithSize `json:"ram"`
}
type NetMetrics struct {
	NetworkName string  `json:"networkName"`
	Value       float64 `json:"value"`
}

type NodeMetricsOptions struct {
	NodeID    string
	Instant   bool
	StartTime string
	EndTime   string
	Step      int
}
