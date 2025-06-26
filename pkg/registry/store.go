package registry

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Store 是我们对持久化层的核心抽象接口。
// 它的方法是通用的，可以处理任何满足 runtime.Object 接口的 API 对象。
type Store interface {
	// Create 将任何 API 对象存入存储。
	Create(obj runtime.Object) error

	// Update 更新一个已存在的对象。
	Update(obj runtime.Object) error

	// Get 将一个指定类型的对象读取出来，并填充到 objInto 中。
	Get(namespace, name string, objInto runtime.Object) error

	// List 列出指定命名空间下某一类型的所有对象，并填充到 listInto 中。
	List(namespace string, listInto runtime.Object) error

	// Delete 删除一个指定类型的对象。
	Delete(namespace, name string, objToDelete runtime.Object) error
}
