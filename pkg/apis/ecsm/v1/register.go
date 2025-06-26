// file: pkg/apis/ecsm/v1/register.go

package v1

import (
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
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ECSMService{},
		&ECSMServiceList{},
	)
	return nil
}
