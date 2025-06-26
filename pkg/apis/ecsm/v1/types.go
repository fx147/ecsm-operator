package v1

import metav1 "github.com/fx147/ecsm-operator/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ECSMService 代表一个ECSM服务实例，是ECSM平台上一个无状态应用的核心抽象
type ECSMService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ECSMServiceSpec   `json:"spec,omitempty"`
	Status ECSMServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ECSMServiceList 包含 ECSMService 的列表
type ECSMServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ECSMService `json:"items"`
}

// ECSMServiceSpec 定义了ECSM服务的期望状态
type ECSMServiceSpec struct {
	// 定义了服务的部署策略，决定了容器实例如何分布在节点上
	// +required
	DeploymentStrategy DeploymentStrategy `json:"deploymentStrategy"`

	// 定义了当镜像更新时服务的升级策略
	// +optional
	UpgradeStrategy UpgradeStrategy `json:"upgradeStrategy,omitempty"`

	// Template 是创建新容器实例的关键模版
	// +required
	Template ContainerTemplateSpec `json:"template"`
}

// ECSMServiceStatus 定义了 ECSMService 的状态
type ECSMServiceStatus struct {
	// Replicas 是在 ECSM 平台上实际找到的、属于此服务的容器实例总数。
	// 从查询 API 的 `factor` 字段获取。
	Replicas int32 `json:"replicas"`

	// ReadyReplicas 是当前处于在线且运行中的容器实例数量。
	// 从查询 API 的 `instanceOnline` 字段获取。
	ReadyReplicas int32 `json:"readyReplicas"`

	// ObservedGeneration 是控制器最近一次处理的 ECSMService.metadata.generation。
	// 这对于区分 spec 变更前后的状态非常重要。
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions 提供了标准的机制来报告服务的当前状态。
	// 例如，"Available", "Progressing", "Degraded"。
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// UnderlyingServiceID 是在 ECSM 平台中对应的真实服务 ID。
	// 这对于调试和直接与 ECSM API 交互非常有用。
	// 从查询 API 的 `id` 字段获取。
	// +optional
	UnderlyingServiceID string `json:"underlyingServiceID,omitempty"`
}

type DeploymentStrategyType string

const (
	DeploymentStrategyTypeStatic  DeploymentStrategyType = "Static"
	DeploymentStrategyTypeDynamic DeploymentStrategyType = "Dynamic"
)

// DeploymentStrategy 定义了服务的部署策略，即节点选择策略
type DeploymentStrategy struct {
	// Type 表示部署类型
	// Static：在 `nodes` 字段中指定的每个节点上都部署一个实例。
	// Dynamic：在 `nodePool` 提供的节点池中，部署 `replicas` 个实例。
	// +kubebuilder:validation:Enum=Static;Dynamic
	// +required
	Type DeploymentStrategyType `json:"type"`

	// Replicas 表示动态选择时的指定副本数量
	// 在 Static 策略下此字段被忽略
	// kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Nodes 是在静态策略下指定的节点列表
	// +optional
	// TODO: 其实需要指定Node类型
	Nodes []string `json:"nodes,omitempty"`

	// NodePool 是在动态策略下指定的节点池
	// +optional
	NodePool []string `json:"nodePool,omitempty"`
}

type UpgradeStrategyType string

const (
	UpgradeStrategyTypeNever  UpgradeStrategyType = "Never"
	UpgradeStrategyTypeLarger UpgradeStrategyType = "Larger"
	UpgradeStrategyTypeAlways UpgradeStrategyType = "Always"
)

// UpgradeStrategy 定义了服务的升级策略，即容器镜像更新策略
type UpgradeStrategy struct {
	// Type 对应 ECSM 的 autoUpgrade 字段。
	// "Never": 从不自动更新。
	// "Larger": 当有更高版本的镜像时更新。
	// "Always": 只要有新镜像就更新。
	// 默认为 "Never"。
	// +kubebuilder:validation:Enum=Never;Larger;Always
	// +optional
	Type UpgradeStrategyType `json:"type,omitempty"`
}

type ImagePullPolicyType string

const (
	ImagePullPolicyAlways       ImagePullPolicyType = "Always"
	ImagePullPolicyIfNotPresent ImagePullPolicyType = "IfNotPresent"
	ImagePullPolicyNever        ImagePullPolicyType = "Never"
)

// ContainerTemplateSpec 定义了容器模版
type ContainerTemplateSpec struct {
	// Image 是要运行的容器镜像引用，格式为 "name@tag"。
	// 例如: "njust@1.1"。
	// +required
	Image string `json:"image"`

	// ImagePullPolicy 定义了镜像拉取策略。默认为 "IfNotPresent"。
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +optional
	ImagePullPolicy ImagePullPolicyType `json:"imagePullPolicy,omitempty"`

	// PrePull 定义了是否开启镜像预热，开启后将在部署时向所有节点同步镜像
	// 默认为 False
	// +optional
	Prepull bool `json:"prepull,omitempty"`

	// Hostname 定义了容器的主机名。如果为空，控制器将默认使用服务名称。
	// +optional
	Hostname string `json:"hostname,omitempty"`

	// Command 是容器的入口点。如果为空，则使用镜像默认的入口点。
	// +optional
	Command []string `json:"command,omitempty"`

	// Env 是要注入到容器中的环境变量列表。
	// +optional
	Env []EnvVar `json:"env,omitempty"`

	// Resources 定义了容器的资源请求和限制。
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// VolumeMounts 是要挂载到容器中的卷列表。
	// +optional
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`

	// VSOA 包含了所有与 VSOA 服务相关的配置。
	// +optional
	VSOA *VSOASpec `json:"vsoa,omitempty"`

	// PlatformSpecific 是一个"逃生舱口"，用于设置平台特有的、不常用的底层配置。
	// 普通用户通常不需要关心此部分。
	// +optional
	PlatformSpecific *PlatformSpecificConfig `json:"platformSpecific,omitempty"`
}

// EnvVar 代表一个环境变量
type EnvVar struct {
	// Name 是环境变量的名称。
	Name string `json:"name"`
	// Value 是环境变量的值。
	Value string `json:"value"`
}

type ResourceType string

const (
	ResourceTypeMemory ResourceType = "memory"
	ResourceTypeDisk   ResourceType = "disk"
)

// 对ECSM的资源模型进行了简化和抽象
type ResourceRequirements struct {
	// Limits 定义了容器的资源限制，内存和硬盘。
	// CPU 优先级请通过高级配置进行设置
	// +optional
	Limits map[ResourceType]string `json:"limits,omitempty"`
}

// VolumeMount 定义了共享库的挂载点
type VolumeMount struct {
	// Name 是挂载点的名称
	Name string `json:"name"`

	// HostPath 是主机上的路径，容器将在此路径下挂载卷。
	HostPath string `json:"hostPath"`

	// ContainerPath 是容器内的目标路径
	ContainerPath string `json:"containerPath"`

	// ReadOnly 如果为 true，容器将以只读模式挂载卷。
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
}

// VSOASpec 定义了 VSOA 服务的配置
type VSOASpec struct {
	// Password 是 VSOA 服务的密码
	// +optional
	Password string `json:"password,omitempty"`
	// Port 是 VSOA 监听的端口
	// 如果为0.表示由ECSM动态分配
	// +optional
	Port *int32 `json:"port,omitempty"`
	// HealthCheck 定义了容器的健康检查配置
	// +optional
	HealthCheck *HealthCheckSpec `json:"healthCheck,omitempty"`
}

type HealthCheckSpec struct {
	// InitialDelaySeconds 是健康检查的初始延迟时间，单位为秒
	// +optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
	// TimeoutSeconds 是健康检查的超时时间，单位为秒
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
	// PeriodSeconds 是健康检查的周期时间，单位为秒
	// +optional
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// FailureThreshold 是健康检查失败的阈值，连续失败多少次后将容器视为不健康
	// +optional
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

type ActionType string

const (
	ActionTypeRun  ActionType = "Run"
	ActionTypeLoad ActionType = "Load"
)

type PlatformSpecificConfig struct {
	// Action 定义了容器启动类型
	// run 表示创建并启动
	// load 表示只创建
	Action ActionType `json:"action,omitempty"`
	// Root 是容器运行的根文件系统路径
	Root *RootSpec `json:"root,omitempty"`
	// Platform 定义了容器的平台类型
	Platform *PlatformSpec `json:"platform,omitempty"`
	// SylixOS 包含了所有针对 SylixOS 的底层配置
	// +optional
	SylixOS *SylixOSConfig `json:"sylixos,omitempty"`
}

type SylixOSConfig struct {
	// Devices 定义了设备信息
	// +optional
	Devices []Device `json:"devices,omitempty"`
	// Network 暂时只定义了是否启动FTPD服务器和TELNETD服务器
	// +optional
	Network *NetworkSpec `json:"network,omitempty"`
	// CPU 相关的底层配置
	CPU *SylixOSCPUConfig `json:"cpu,omitempty"`
	// Memory 相关的底层配置
	Memory *SylixOSMemoryConfig `json:"memory,omitempty"`
}

// SylixOSMemoryConfig 包含专属于 SylixOS 的、不常用的内存配置。
// 注意：常用的内存限制（memoryLimitMB）应该通过 spec.template.resources.limits.memory 来设置。
type SylixOSMemoryConfig struct {
	// KheapLimit 是内核堆限制（字节）。
	// 这是一个高级设置，大多数用户不需要关心。
	// 如果用户未指定，控制器可以应用 ECSM 的默认值。
	// +optional
	KheapLimit *int64 `json:"kheapLimit,omitempty"`
}

// SylixOSCPUConfig 包含专属于 SylixOS 的 CPU 优先级设置
// TODO：通过这种设置CPU优先级的方式对用户是否有些麻烦，如果用户设置优先级的操作比较频繁，需要考虑转移到顶层
type SylixOSCPUConfig struct {
	// HighestPrio 是最高优先级
	// +optional
	HighestPrio *int64 `json:"highestPrio,omitempty"`
	// LowestPrio 是最低优先级
	// +optional
	LowestPrio *int64 `json:"lowestPrio,omitempty"`
}

type NetworkSpec struct {
	// FTPD 定义了是否启动FTPD服务器
	// +optional
	FTPD bool `json:"ftpd,omitempty"`
	// TELNETD 定义了是否启动TELNETD服务器
	// +optional
	TELNETD bool `json:"telnetd,omitempty"`
}

type Device struct {
	// 文件路径
	Path string `json:"path"`
	// 访问权限
	// TODO:这里暂时不清楚具体三种权限的英文
	Access string `json:"access"`
}

type PlatformSpec struct {
	// OS 代表镜像系统
	// +optional
	OS string `json:"os,omitempty"`
	// Arch 代表镜像架构
	// +optional
	Arch string `json:"arch,omitempty"`
}

// RootSpec 定义了容器运行时的根文件系统路径
type RootSpec struct {
	// Path 是容器运行时的根文件系统路径
	Path string `json:"path,omitempty"`
	// ReadOnly 是容器运行时的根文件系统是否只读
	ReadOnly bool `json:"readOnly,omitempty"`
}
