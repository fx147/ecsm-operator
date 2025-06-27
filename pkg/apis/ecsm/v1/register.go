// file: pkg/apis/ecsm/v1/register.go

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupName 是我们 API Group 的名称
const GroupName = "ecsm.sh"

// SchemeGroupVersion is group version used to register these objects.
// 这就是我们缺失的、可供外部使用的变量。
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// addKnownTypes adds the known types to the Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	// 这里注册自己定义的核心资源
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ECSMService{},
		&ECSMServiceList{},
	)

	// 这里注册通用的辅助性的元数据类型
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)

	return nil
}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind.
// Kind 接收一个不带 Group 的 Kind (例如 "ECSMService")，
// 并返回一个包含了我们 API Group 的、完整的 GroupKind。
//
// 使用示例: metav1.Kind("ECSMService")
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource.
// Resource 接收一个不带 Group 的 Resource (例如 "ecsmservices")，
// 并返回一个包含了我们 API Group 的、完整的 GroupResource。
//
// 使用示例: metav1.Resource("ecsmservices")
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
