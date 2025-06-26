// file: main.go

package main

import (
	"fmt"
	"reflect"

	v1 "github.com/fx147/ecsm-operator/pkg/apis/ecsm/v1"
	ecsmmetav1 "github.com/fx147/ecsm-operator/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	fmt.Println("--- Starting Interface Implementation Test ---")

	// --- 1. 编译时检查 ---
	// 这是最关键的测试。我们声明一个 runtime.Object 类型的变量，
	// 然后尝试将一个 *v1alpha1.ECSMService 的零值赋给它。
	// 如果这行代码能够编译通过，就证明 *ECSMService 已经完全实现了 runtime.Object 接口。
	// 如果有任何方法缺失，这里就会产生一个编译错误。
	var _ runtime.Object = &v1.ECSMService{}
	fmt.Println("[SUCCESS] Compile-time check passed: ECSMService implements runtime.Object.")

	fmt.Println("\n--- Running Runtime Method Tests ---")

	// 创建一个具体的 ECSMService 实例用于测试
	originalService := &v1.ECSMService{
		TypeMeta: ecsmmetav1.TypeMeta{
			APIVersion: "ecsm.sh/v1alpha1",
			Kind:       "ECSMService",
		},
		ObjectMeta: ecsmmetav1.ObjectMeta{
			Name:      "my-test-app",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
	}

	// --- 2. 运行时检查 GetObjectKind() ---
	gvk := originalService.GetObjectKind().GroupVersionKind()
	fmt.Printf("[INFO] Called GetObjectKind(), got GVK: %s\n", gvk.String())
	if gvk.Kind == "ECSMService" && gvk.Group == "ecsm.sh" {
		fmt.Println("[SUCCESS] GetObjectKind() returned correct Kind and Group.")
	} else {
		fmt.Println("[FAILURE] GetObjectKind() returned unexpected GVK.")
	}

	// --- 3. 运行时检查 DeepCopyObject() ---
	fmt.Println("\n[INFO] Calling DeepCopyObject()...")
	copiedObject := originalService.DeepCopyObject()

	// Type Assertion: DeepCopyObject() 返回的是 runtime.Object 接口类型，
	// 我们需要将它转换回具体的 *v1alpha1.ECSMService 类型才能访问其字段。
	copiedService, ok := copiedObject.(*v1.ECSMService)
	if !ok {
		fmt.Println("[FAILURE] DeepCopyObject() returned an object of unexpected type.")
		return
	}
	fmt.Println("[SUCCESS] Type assertion from runtime.Object to *ECSMService succeeded.")

	// 验证深拷贝的有效性
	// 检查1: 内存地址不同，证明是两个独立的对象
	if &originalService != &copiedService {
		fmt.Println("[SUCCESS] Pointers of original and copied objects are different (correct).")
	} else {
		fmt.Println("[FAILURE] Pointers are the same, this is a shallow copy!")
	}

	// 检查2: 内容相同
	if reflect.DeepEqual(originalService.Spec, copiedService.Spec) &&
		reflect.DeepEqual(originalService.ObjectMeta, copiedService.ObjectMeta) {
		fmt.Println("[SUCCESS] Contents of original and copied objects are equal (correct).")
	} else {
		fmt.Println("[FAILURE] Contents are not equal.")
	}

	// 检查3: 修改副本，不影响原始对象 (深拷贝的终极证明)
	copiedService.ObjectMeta.Labels["app"] = "modified"
	fmt.Printf("[INFO] Modified copied object's label to: %s\n", copiedService.ObjectMeta.Labels["app"])
	fmt.Printf("[INFO] Original object's label remains: %s\n", originalService.ObjectMeta.Labels["app"])
	if originalService.ObjectMeta.Labels["app"] == "test" {
		fmt.Println("[SUCCESS] Modifying the copy did NOT affect the original object.")
	} else {
		fmt.Println("[FAILURE] Modifying the copy affected the original object!")
	}

	fmt.Println("\n--- All Tests Passed! ---")
}
