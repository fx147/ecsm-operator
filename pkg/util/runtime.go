package util

import (
	"fmt"
	"reflect"

	metav1 "github.com/fx147/ecsm-operator/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ObjectGVKAndMeta 是一个辅助函数，用于从 runtime.Object 中提取 GVK 和元数据。
// 这是实现通用存储和控制器的关键。
func ObjectGVKAndMeta(obj runtime.Object) (gvk schema.GroupVersionKind, meta *metav1.ObjectMeta, err error) {
	gvk = obj.GetObjectKind().GroupVersionKind()

	// 使用反射来安全地获取 ObjectMeta 字段
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return gvk, nil, fmt.Errorf("object is not a struct")
	}

	metaField := val.FieldByName("ObjectMeta")
	if !metaField.IsValid() {
		return gvk, nil, fmt.Errorf("object does not have ObjectMeta field")
	}

	meta, ok := metaField.Addr().Interface().(*metav1.ObjectMeta)
	if !ok {
		return gvk, nil, fmt.Errorf("field ObjectMeta is not of type *metav1.ObjectMeta")
	}

	return gvk, meta, nil
}