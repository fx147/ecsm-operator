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

// --- 辅助函数 (已重构) ---

// getGVK 根据对象的 Go 类型返回其 GroupVersionKind。
// 它可以安全地处理单数和列表类型的对象。
func GetGVK(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("failed to get object kinds: %w", err)
	}
	if len(gvks) == 0 {
		return schema.GroupVersionKind{}, fmt.Errorf("no object kinds found")
	}
	// gvk 可能返回多个
	gvk := gvks[0]

	return gvk, nil
}

// getObjectMeta 从对象中提取 ObjectMeta。
// 如果对象没有 ObjectMeta 字段，它将返回错误。
func GetObjectMeta(obj runtime.Object) (*metav1.ObjectMeta, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object is not a struct")
	}

	metaField := val.FieldByName("ObjectMeta")
	if !metaField.IsValid() {
		return nil, fmt.Errorf("object does not have ObjectMeta field")
	}

	meta, ok := metaField.Addr().Interface().(*metav1.ObjectMeta)
	if !ok {
		return nil, fmt.Errorf("field ObjectMeta is not of type *metav1.ObjectMeta")
	}
	return meta, nil
}
