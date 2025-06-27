// file: pkg/registry/filestore_test.go

package registry

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	ecsmv1 "github.com/fx147/ecsm-operator/pkg/apis/ecsm/v1"
	metav1 "github.com/fx147/ecsm-operator/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// newTestService 是一个辅助函数，用于快速创建一个测试用的 ECSMService 对象。
func newTestService(namespace, name string) *ecsmv1.ECSMService {
	return &ecsmv1.ECSMService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ecsmv1.SchemeGroupVersion.String(),
			Kind:       "ECSMService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    map[string]string{"app": name},
		},
		Spec: ecsmv1.ECSMServiceSpec{
			// ...可以填充一些简单的 spec 用于测试
		},
	}
}

// newTestScheme 是一个新的辅助函数，专门用于创建和初始化我们的 Scheme。
// 这让我们的测试代码更清晰。
func newTestScheme() *runtime.Scheme {
	// 1. 创建一个全新的、空的 Scheme 实例。
	s := runtime.NewScheme()

	// 2. 调用我们 API 包中 register.go 文件提供的 AddToScheme 函数。
	//    这将把 ECSMService 和 ECSMServiceList 的类型信息注册到 s 中。
	//    我们在这里可以忽略错误，因为我们确信它会成功。
	_ = ecsmv1.AddToScheme(s)

	return s
}

func TestFileStore(t *testing.T) {
	// 1. 创建一个临时的测试目录
	// t.TempDir() 是 Go 1.15+ 的一个非常有用的功能，它会在测试结束后自动清理目录。
	tempDir := t.TempDir()

	// 2. 创建新的Scheme
	testScheme := newTestScheme()

	// 3. 创建 FileStore 实例
	store, err := NewFileStore(tempDir, testScheme)
	if err != nil {
		t.Fatalf("Failed to create FileStore: %v", err)
	}

	// 2. 定义我们的测试用例
	ns1 := "default"
	ns2 := "production"
	svc1 := newTestService(ns1, "app-one")
	svc2 := newTestService(ns1, "app-two")
	svc3 := newTestService(ns2, "app-one") // 在另一个命名空间下的同名服务

	// --- 测试 Create 和 Get ---
	t.Run("CreateAndGet", func(t *testing.T) {
		// 创建 svc1
		if err := store.Create(svc1); err != nil {
			t.Fatalf("Create(svc1) failed: %v", err)
		}

		// 尝试获取刚创建的 svc1
		retrievedSvc1 := &ecsmv1.ECSMService{}
		if err := store.Get(ns1, "app-one", retrievedSvc1); err != nil {
			t.Fatalf("Get(svc1) failed: %v", err)
		}

		// 深度比较，确保取回来的对象和存进去的一样
		if !reflect.DeepEqual(svc1, retrievedSvc1) {
			t.Errorf("Retrieved object is not equal to the original object. Got %+v, want %+v", retrievedSvc1, svc1)
		}
	})

	// --- 测试重复创建 ---
	t.Run("CreateAlreadyExists", func(t *testing.T) {
		err := store.Create(svc1)
		if !errors.IsAlreadyExists(err) {
			t.Errorf("Expected 'AlreadyExists' error, but got: %v", err)
		}
	})

	// --- 测试获取不存在的对象 ---
	t.Run("GetNotFound", func(t *testing.T) {
		nonExistentSvc := &ecsmv1.ECSMService{}
		err := store.Get(ns1, "non-existent", nonExistentSvc)
		if !errors.IsNotFound(err) {
			t.Errorf("Expected 'NotFound' error, but got: %v", err)
		}
	})

	// --- 测试 List ---
	t.Run("List", func(t *testing.T) {
		// 创建更多对象用于测试 List
		if err := store.Create(svc2); err != nil {
			t.Fatalf("Create(svc2) failed: %v", err)
		}
		if err := store.Create(svc3); err != nil {
			t.Fatalf("Create(svc3) failed: %v", err)
		}

		// 列出 ns1 (default) 命名空间下的服务
		list1 := &ecsmv1.ECSMServiceList{}
		if err := store.List(ns1, list1); err != nil {
			t.Fatalf("List in namespace '%s' failed: %v", ns1, err)
		}
		if len(list1.Items) != 2 {
			t.Errorf("Expected 2 items in namespace '%s', but got %d", ns1, len(list1.Items))
		}

		// 列出 ns2 (production) 命名空间下的服务
		list2 := &ecsmv1.ECSMServiceList{}
		if err := store.List(ns2, list2); err != nil {
			t.Fatalf("List in namespace '%s' failed: %v", ns2, err)
		}
		if len(list2.Items) != 1 {
			t.Errorf("Expected 1 item in namespace '%s', but got %d", ns2, len(list2.Items))
		}
	})

	// --- 测试 Update ---
	t.Run("Update", func(t *testing.T) {
		// 获取 svc2，修改它的标签，然后更新
		svcToUpdate := &ecsmv1.ECSMService{}
		if err := store.Get(ns1, "app-two", svcToUpdate); err != nil {
			t.Fatalf("Get for update failed: %v", err)
		}

		svcToUpdate.ObjectMeta.Labels["updated"] = "true"
		if err := store.Update(svcToUpdate); err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// 再次获取，验证修改已生效
		updatedSvc := &ecsmv1.ECSMService{}
		if err := store.Get(ns1, "app-two", updatedSvc); err != nil {
			t.Fatalf("Get after update failed: %v", err)
		}
		if updatedSvc.ObjectMeta.Labels["updated"] != "true" {
			t.Errorf("Update was not persisted. Expected label 'updated:true', but it was not found.")
		}
	})

	// --- 测试 Delete ---
	t.Run("Delete", func(t *testing.T) {
		// 删除 svc1
		if err := store.Delete(ns1, "app-one", &ecsmv1.ECSMService{}); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// 再次获取，应该得到 NotFound 错误
		err := store.Get(ns1, "app-one", &ecsmv1.ECSMService{})
		if !errors.IsNotFound(err) {
			t.Errorf("Expected 'NotFound' error after delete, but got: %v", err)
		}

		// 删除一个不存在的对象，不应该报错
		if err := store.Delete(ns1, "non-existent", &ecsmv1.ECSMService{}); err != nil {
			t.Errorf("Deleting a non-existent object should not return an error, but got: %v", err)
		}
	})

	// 验证最终文件结构是否正确（可选，用于调试）
	t.Run("VerifyFileStructure", func(t *testing.T) {
		expectedPath := filepath.Join(tempDir, "ecsm.sh", "v1", "ecsmservices", "production", "app-one.json")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Expected file to exist at %s, but it was not found.", expectedPath)
		}
	})
}
