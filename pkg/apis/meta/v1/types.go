package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// 描述了资源的类型
type TypeMeta struct {
	// 此对象所表示的REST资源
	// +required
	Kind string `json:"kind,omitempty"`

	// 定义了此对象表示的版本，例如"apps/v1"
	// +required
	APIVersion string `json:"apiVersion,omitempty"`
}

// 描述一个资源实例所需要的元数据
type ObjectMeta struct {
	// 用户通过yaml文件创建资源实例时指定的名称
	// +required
	Name string `json:"name"`

	// 资源所属的命名空间，现在ECSM并没有命名空间的概念，默认为default
	// 可以硬编码，保留字段以备将来扩展
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// 资源实例的唯一标识符，由系统自动生成
	// +readonly
	UID string `json:"uid,omitempty"`

	// 用于筛选和选择对象的标签键值对
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// 用于附加任意非标识性元数据的键值对。通常用于存储工具、库或人的信息，不用于选择对象
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// 内部版本号，用于实现乐观并发控制。当资源被更新时，此值会递增。客户端可以基于此值来判断资源是否已被修改
	// TODO:数据库不存这个值，感觉实现起来比较困难
	// +readonly
	ResourceVersion string `json:"resourceVersion,omitempty"`

	// 创建时间
	// +readonly
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty"`
	// 删除时间，如果不为nil，表示对象正在被删除
	// +readonly
	DeletionTimestamp *metav1.Time `json:"deletionTimestamp,omitempty"`
}

// ListMeta 包含了列表（集合）资源所需的元数据。
// 主要用于支持分页和集合的版本控制。
type ListMeta struct {
	// ResourceVersion 是一个字符串，表示此列表所代表的资源的版本。
	// 客户端可以用它来发起 watch 请求。
	// 在项目的早期阶段可以暂时不填充这个字段，但定义它以保持 API 兼容性。
	// +optional
	ResourceVersion string `json:"resourceVersion,omitempty"`

	// Continue 是一个不透明的令牌，用于从服务器获取下一页的结果。
	// 如果此字段为空，表示没有更多页。
	// 同样，在早期可以不实现分页功能，但保留此字段。
	// +optional
	Continue string `json:"continue,omitempty"`

	// RemainingItemCount 是指在当前分页请求之后，仍然剩余的项目数量。
	// +optional
	RemainingItemCount *int64 `json:"remainingItemCount,omitempty"`
}

// ConditionStatus 是 Condition的状态
type ConditionStatus string

const (
	ConditionStatusTrue    ConditionStatus = "True"
	ConditionStatusFalse   ConditionStatus = "False"
	ConditionStatusUnknown ConditionStatus = "Unknown"
)

type Condition struct {
	// Type 是condition的类型，例如Ready
	// +required
	Type string `json:"type,omitempty"`
	// Status 是condition的状态
	// +required
	Status ConditionStatus `json:"status,omitempty"`
	// LastTransitionTime 是condition最后一次转换的时间
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason 是condition转换的原因
	// +optional
	Reason string `json:"reason,omitempty"`
	// Message 是人类可读的详细信息
	// +optional
	Message string `json:"message,omitempty"`
}

type OwnerReference struct {
	// APIVersion 是引用对象的API版本
	// +required
	APIVersion string `json:"apiVersion,omitempty"`
	// Kind 是引用对象的资源类型
	// +required
	Kind string `json:"kind,omitempty"`
	// Name 是引用对象的名称
	// +required
	Name string `json:"name,omitempty"`
	// UID 是引用对象的唯一标识符
	// +required
	UID string `json:"uid,omitempty"`
	// Controller 是一个布尔值，表示是否是控制器
	// +optional
	Controller *bool `json:"controller,omitempty"`
}

// GetObjectKind 返回一个指向该对象类型信息的指针。
// 因为 *TypeMeta 实现了 schema.ObjectKind 接口，所以可以直接返回自身。
func (t *TypeMeta) GetObjectKind() schema.ObjectKind {
	return t
}

// SetGroupVersionKind 为对象设置 GroupVersionKind 信息。
// 这是实现 schema.ObjectKind 接口所必需的方法。
func (t *TypeMeta) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	t.APIVersion, t.Kind = gvk.ToAPIVersionAndKind()
}

// GroupVersionKind 返回对象的 GroupVersionKind。
// 如果 APIVersion 或 Kind 为空，它可能返回不完整的 GVK。
// 这也是实现 schema.ObjectKind 接口所必需的方法。
func (t *TypeMeta) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(t.APIVersion, t.Kind)
}
