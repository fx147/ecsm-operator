// file: pkg/ecsm_client/clientset/container_types.go

package clientset

// --- Container Get && List Structures ---

// ContainerInfo 精确映射了 ECSM API 中 Container 对象的 JSON 结构。
// 它同时用于 List 和 Get 的响应。
type ContainerInfo struct {
	ID              string   `json:"id"`
	TaskID          string   `json:"taskId"`
	Name            string   `json:"name"`
	Status          string   `json:"status"`
	Uptime          int      `json:"uptime"`
	StartedTime     string   `json:"startedTime"`
	CreatedTime     string   `json:"createdTime"`
	TaskCreatedTime string   `json:"taskCreatedTime"`
	DeployStatus    string   `json:"deployStatus"`
	FailedMessage   *string  `json:"failedMessage"` // Can be null
	RestartCount    int      `json:"restartCnt"`
	DeployNum       int      `json:"deployNum"`
	CPUUsage        CPUUsage `json:"cpuUsage"`
	MemoryLimit     int64    `json:"memoryLimit"`
	MemoryUsage     int64    `json:"memoryUsage"`
	MemoryMaxUsage  int64    `json:"memoryMaxUsage"`
	SizeUsage       int64    `json:"sizeUsage"`
	SizeLimit       int64    `json:"sizeLimit"`
	ServiceID       string   `json:"serviceId"`
	ServiceName     string   `json:"serviceName"`
	NodeID          string   `json:"nodeId"`
	Address         string   `json:"address"`
	NodeName        string   `json:"nodeName"`
	NodeArch        string   `json:"nodeArch"`
	ImageID         string   `json:"imageId"`
	ImageName       string   `json:"imageName"`
	ImageVersion    string   `json:"imageVersion"`
	ImageOS         string   `json:"imageOS"`
	ImageArch       string   `json:"imageArch"`
}

// CPUUsage 描述了容器的 CPU 使用情况。
type CPUUsage struct {
	Total float64   `json:"total"`
	Cores []float64 `json:"cores"`
}

// ContainerList 是 ListByService 和 ListByNode 方法的返回值。
type ContainerList struct {
	Total    int             `json:"total"`
	PageNum  int             `json:"pageNum"`
	PageSize int             `json:"pageSize"`
	Items    []ContainerInfo `json:"list"`
}

// ListContainersByServiceOptions 封装了查询服务下容器列表的参数。
type ListContainersByServiceOptions struct {
	PageNum    int      `json:"pageNum"`
	PageSize   int      `json:"pageSize"`
	ServiceIDs []string `json:"serviceIds"` // 必填
	Key        string   `json:"key,omitempty"`
}

type ListContainersByNodeOptions struct {
	PageNum  int      `json:"pageNum"`
	PageSize int      `json:"pageSize"`
	NodeIDs  []string `json:"nodeIds"` // 必填
	Key      string   `json:"key,omitempty"`
}

// --- Container Control Structures ---

// ContainerAction 定义了可以对容器执行的动作类型。
type ContainerAction string

const (
	ActionStart   ContainerAction = "start"
	ActionStop    ContainerAction = "stop"
	ActionRestart ContainerAction = "restart"
	ActionPause   ContainerAction = "pause"
	ActionUnpause ContainerAction = "unpause"
)

// ContainerControlRequest 定义了控制容器状态的 API payload。
type ContainerControlByNameRequest struct {
	// API 字段是 "id"，但含义是 name
	Name   string          `json:"id"`
	Action ContainerAction `json:"action"`
}

type ServiceControlContainerRequest struct {
	ID     string          `json:"serviceId"`
	Action ContainerAction `json:"action"`
}

// Transaction 描述了一个异步操作任务。
type Transaction struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"` // "running", "failure", "success"
	Data      interface{} `json:"data"`   // 使用 interface{} 来匹配任意对象
	Timestamp int64       `json:"timestamp"`
}

// --- Container History Structures ---

// ContainerHistoryOptions 封装了查询容器操作历史的参数。
type ContainerHistoryOptions struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
	// 注意：API文档中的 'id' 字段指的是 Task ID。
	TaskID string `json:"id"`
}

// ContainerHistoryList 是 GetHistory 方法的返回值。
type ContainerHistoryList struct {
	Total    int                `json:"total"`
	PageNum  int                `json:"pageNum"`
	PageSize int                `json:"pageSize"`
	Items    []ContainerHistory `json:"list"`
}

// ContainerHistory 代表单条容器操作历史记录。
type ContainerHistory struct {
	ID   string `json:"id"`
	Cmd  string `json:"cmd"`
	User string `json:"user"`
	Time string `json:"time"`
}
