package clientset

type EcsImageConfig struct {
	Platform *Platform `json:"platform,omitempty"`
	Process  *Process  `json:"process,omitempty"`
	Root     *Root     `json:"root,omitempty"`
	Hostname string    `json:"hostname,omitempty"`
	Mounts   []Mount   `json:"mounts,omitempty"`
	SylixOS  *SylixOS  `json:"sylixos,omitempty"`
}

type Platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type Process struct {
	Args []string `json:"args"`
	Env  []string `json:"env"`
	Cwd  string   `json:"cwd"`
}

type Root struct {
	Path     string `json:"path"`
	Readonly bool   `json:"readonly"`
}

type Mount struct {
	Destination string   `json:"destination"`
	Source      string   `json:"source"`
	Options     []string `json:"options"`
}

type SylixOS struct {
	Devices   []Device   `json:"devices,omitempty"`
	Resources *Resources `json:"resources"`
	Network   *Network   `json:"network"`
	Commands  []string   `json:"commands"`
}

type Device struct {
	Path   string `json:"path"`
	Access string `json:"access"`
}

type Resources struct {
	CPU          *CPU          `json:"cpu"`
	Memory       *Memory       `json:"memory"`
	Disk         *Disk         `json:"disk"`
	KernelObject *KernelObject `json:"kernelObject"`
}

type CPU struct {
	HighestPrio int `json:"highestPrio"`
	LowestPrio  int `json:"lowestPrio"`
}

type Memory struct {
	KheapLimit    int `json:"kheapLimit"`
	MemoryLimitMB int `json:"memoryLimitMB"`
}

type Disk struct {
	LimitMB int `json:"limitMB"`
}

type KernelObject struct {
	ThreadLimit        int `json:"threadLimit"`
	ThreadPoolLimit    int `json:"threadPoolLimit"`
	EventLimit         int `json:"eventLimit"`
	EventSetLimit      int `json:"eventSetLimit"`
	PartitionLimit     int `json:"partitionLimit"`
	RegionLimit        int `json:"regionLimit"`
	MsgQueueLimit      int `json:"msgQueueLimit"`
	TimerLimit         int `json:"timerLimit"`
	RMSLimit           int `json:"rmsLimit,omitempty"` // API 文档中没有，但 payload 示例中有
	ThreadVarLimit     int `json:"threadVarLimit,omitempty"`
	PosixMqueueLimit   int `json:"posixMqueueLimit,omitempty"`
	DlopenLibraryLimit int `json:"dlopenLibraryLimit,omitempty"`
	XSIIPCLimit        int `json:"xsiipcLimit,omitempty"`
	SocketLimit        int `json:"socketLimit,omitempty"`
	SRTPLimit          int `json:"srtpLimit,omitempty"`
	DeviceLimit        int `json:"deviceLimit,omitempty"`
}

type Network struct {
	FtpdEnable    bool `json:"ftpdEnable"`
	TelnetdEnable bool `json:"telnetdEnable"`
}

// ImageListOptions 封装了所有可以用于 List 镜像的查询参数。
type ImageListOptions struct {
	// RegistryID 是要查询的仓库主键，本地仓库为 "local"。
	RegistryID string // 必填
	PageNum    int
	PageSize   int
	Name       string
	OS         string
	Author     string
}

// ImageList 是 List 方法的返回值，精确匹配 API 响应中的 data 字段。
type ImageList struct {
	Total    int             `json:"total"`
	PageNum  int             `json:"pageNum"`
	PageSize int             `json:"pageSize"`
	Items    []ImageListItem `json:"list"`
}

// ImageListItem 代表镜像列表中的单个镜像信息。
type ImageListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	OS          string  `json:"os"`
	CreatedTime string  `json:"createdTime"`
	Tag         string  `json:"tag"`
	Size        float64 `json:"size"`
	Author      *string `json:"author"` // author 可以为 null，使用指针
	Arch        string  `json:"arch"`
	Pulled      bool    `json:"pulled"`
	Description string  `json:"description"`
	OCIVersion  string  `json:"ociVersion"`
}

// ImageStatistics 描述了镜像的统计信息，精确匹配 /image/summary API 的响应。
type ImageStatistics struct {
	Local  int `json:"local"`
	Remote int `json:"remote"`
}

// ImageDetails 代表通过 Get Details API 获取到的镜像详细信息。
type ImageDetails struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	OS          string          `json:"os"`
	Arch        string          `json:"arch"`
	CreatedTime string          `json:"createdTime"`
	Size        float64         `json:"size"`
	Author      *string         `json:"author"`
	RawConfig   string          `json:"rawConfig"`
	Config      *EcsImageConfig `json:"config"`
	OCIVersion  string          `json:"ociVersion"`
	Hostname    *string         `json:"hostname"`
	Tag         string          `json:"tag"`
	Pulled      bool            `json:"pulled"`
	Delete      bool            `json:"delete"`
}

// RepositoryInfoOptions 封装了查询镜像仓库信息时的过滤参数。
type RepositoryInfoOptions struct {
	Name   string
	OS     string
	Author string
}

// RepositoryInfo 描述了单个镜像仓库的统计和状态信息。
type RepositoryInfo struct {
	Count        int    `json:"count"` // 响应示例中有，但文档没有，我们加上
	RegistryID   string `json:"registryId"`
	RegistryName string `json:"registryName"`
	// status 和 standard 是可选的，因为 local 仓库没有这两个字段
	Status   *bool `json:"status,omitempty"`
	Standard *bool `json:"standard,omitempty"`
}
