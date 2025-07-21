package rest

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// ExampleRESTClient 演示如何使用 REST 客户端获取服务列表
// 这是一个示例函数，展示了基本的使用方法
func ExampleRESTClient() {
	// 创建 REST 客户端连接到 ECSM API
	client, err := NewRESTClient("http", "192.168.31.129", "3001", &http.Client{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		fmt.Printf("Failed to create REST client: %v\n", err)
		return
	}

	// 构建并执行 GET 请求
	ctx := context.Background()
	result := client.Get().
		Resource("service").
		Param("pageNum", "1").
		Param("pageSize", "10").
		Do(ctx)

	// 解析响应到结构体
	var serviceList ServiceListResponse
	err = result.Into(&serviceList)
	if err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		return
	}

	// 打印结果
	fmt.Printf("Total services: %d\n", serviceList.Total)
	fmt.Printf("Page size: %d\n", serviceList.PageSize)
	fmt.Printf("Page number: %d\n", serviceList.PageNum)
	fmt.Printf("Services found: %d\n", len(serviceList.List))

	for i, service := range serviceList.List {
		fmt.Printf("Service %d:\n", i+1)
		fmt.Printf("  ID: %s\n", service.ID)
		fmt.Printf("  Name: %s\n", service.Name)
		fmt.Printf("  Status: %s\n", service.Status)
		fmt.Printf("  Created: %s\n", service.CreatedTime)
		fmt.Printf("  Updated: %s\n", service.UpdatedTime)
		fmt.Printf("  Policy: %s\n", service.Policy)
		fmt.Printf("  Factor: %d\n", service.Factor)
		fmt.Println()
	}
}

// TestManualRESTClient 手动测试函数，可以用来快速验证与真实 API 的连接
// 运行方式: go test -v -run TestManualRESTClient
func TestManualRESTClient(t *testing.T) {
	// 如果不想运行真实 API 测试，可以跳过
	// t.Skip("Manual test - uncomment to run against real ECSM API")

	// 创建客户端
	client, err := NewRESTClient("http", "192.168.31.129", "3001", &http.Client{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create REST client: %v", err)
	}

	// 测试获取服务列表
	t.Log("Testing GET /api/v1/service")
	ctx := context.Background()
	result := client.Get().
		Resource("service").
		Param("pageNum", "1").
		Param("pageSize", "10").
		Do(ctx)

	var serviceList ServiceListResponse
	err = result.Into(&serviceList)
	if err != nil {
		t.Fatalf("Failed to get services: %v", err)
	}

	t.Logf("Successfully retrieved %d services (total: %d)", len(serviceList.List), serviceList.Total)

	// 如果有服务，打印第一个服务的详细信息
	if len(serviceList.List) > 0 {
		service := serviceList.List[0]
		t.Logf("First service: ID=%s, Name=%s, Status=%s", service.ID, service.Name, service.Status)
	}

	// 测试其他可能的端点（如果需要）
	// 例如：测试获取特定服务的详细信息
	if len(serviceList.List) > 0 {
		serviceID := serviceList.List[0].ID
		t.Logf("Testing GET /api/v1/service/%s", serviceID)

		result = client.Get().
			Resource("service").
			Name(serviceID).
			Do(ctx)

		// 这里可以定义单个服务的响应结构体并解析
		// 暂时只检查是否有错误
		var singleService interface{}
		err = result.Into(&singleService)
		if err != nil {
			t.Logf("Failed to get single service (this might be expected if endpoint doesn't exist): %v", err)
		} else {
			t.Logf("Successfully retrieved single service details")
		}
	}
}

// BenchmarkRESTClient_GetServices 性能测试
func BenchmarkRESTClient_GetServices(b *testing.B) {
	// b.Skip("Benchmark test - uncomment to run performance test against real ECSM API")

	client, err := NewRESTClient("http", "192.168.31.129", "3001", &http.Client{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		b.Fatalf("Failed to create REST client: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := client.Get().
			Resource("service").
			Param("pageNum", "1").
			Param("pageSize", "10").
			Do(ctx)

		var serviceList ServiceListResponse
		err = result.Into(&serviceList)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
	}
}
