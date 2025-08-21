// file: pkg/registry/event.go

package registry

import "k8s.io/apimachinery/pkg/runtime"

// EventType 定义了事件的类型
type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
)

// Event 是一个描述 API 对象变更的事件。
type Event struct {
	Type EventType
	// Key 是对象的唯一标识，例如 "default/my-app"
	Key string
	// Obj 是事件关联的对象
	Object runtime.Object
	// ResourceVersion 是变更后对象的 resourceVersion
	ResourceVersion string
}
