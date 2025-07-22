package rest

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// ServiceListResponse 表示服务列表的响应结构
type ServiceListResponse struct {
	List     []ServiceInfo `json:"list"`
	Total    int           `json:"total"`
	PageSize int           `json:"pageSize"`
	PageNum  int           `json:"pageNum"`
}

// ServiceInfo 表示单个服务的信息
type ServiceInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`
	Status      string `json:"status"`
	Factor      int    `json:"factor"`
	Policy      string `json:"policy"`
}

// TestRESTClient_GetServices 测试获取服务列表的功能
func TestRESTClient_GetServices(t *testing.T) {
	// 创建模拟的 ECSM API 服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/service" {
			t.Errorf("Expected path /api/v1/service, got %s", r.URL.Path)
		}

		// 验证查询参数
		pageNum := r.URL.Query().Get("pageNum")
		pageSize := r.URL.Query().Get("pageSize")
		if pageNum != "1" {
			t.Errorf("Expected pageNum=1, got %s", pageNum)
		}
		if pageSize != "10" {
			t.Errorf("Expected pageSize=10, got %s", pageSize)
		}

		// 返回模拟响应
		mockResponse := map[string]interface{}{
			"status":  200,
			"message": "success",
			"data": map[string]interface{}{
				"list": []map[string]interface{}{
					{
						"id":          "4b0b858f101817174285005970307d414",
						"name":        "acc_server",
						"createdTime": "2024-06-03 23:28:20",
						"updatedTime": "2024-06-03 23:28:23",
						"status":      "complete",
						"factor":      1,
						"policy":      "static",
					},
				},
				"total":    1,
				"pageSize": 10,
				"pageNum":  1,
			},
			"fieldErrors": nil,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// 创建 REST 客户端
	addr := mockServer.Listener.Addr().(*net.TCPAddr)
	client, err := NewRESTClient("http", addr.IP.String(),
		strconv.Itoa(addr.Port), &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Failed to create REST client: %v", err)
	}

	// 执行 GET 请求
	ctx := context.Background()
	result := client.Get().
		Resource("service").
		Param("pageNum", "1").
		Param("pageSize", "10").
		Do(ctx)

	// 解析响应
	var serviceList ServiceListResponse
	err = result.Into(&serviceList)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 验证响应数据
	if serviceList.Total != 1 {
		t.Errorf("Expected total=1, got %d", serviceList.Total)
	}
	if serviceList.PageSize != 10 {
		t.Errorf("Expected pageSize=10, got %d", serviceList.PageSize)
	}
	if serviceList.PageNum != 1 {
		t.Errorf("Expected pageNum=1, got %d", serviceList.PageNum)
	}
	if len(serviceList.List) != 1 {
		t.Errorf("Expected 1 service in list, got %d", len(serviceList.List))
	}

	// 验证服务详细信息
	if len(serviceList.List) > 0 {
		service := serviceList.List[0]
		if service.ID != "4b0b858f101817174285005970307d414" {
			t.Errorf("Expected service ID 4b0b858f101817174285005970307d414, got %s", service.ID)
		}
		if service.Name != "acc_server" {
			t.Errorf("Expected service name acc_server, got %s", service.Name)
		}
		if service.Status != "complete" {
			t.Errorf("Expected service status complete, got %s", service.Status)
		}
	}
}

// TestRESTClient_RealAPI 测试真实的 ECSM API（需要真实的服务器运行）
// 这个测试默认跳过，可以通过 go test -v -run TestRESTClient_RealAPI 单独运行
func TestRESTClient_RealAPI(t *testing.T) {
	//t.Skip("Skipping real API test - uncomment to test against real ECSM server")

	// 创建连接到真实 ECSM API 的客户端
	client, err := NewRESTClient("http", "192.168.31.129", "3001", &http.Client{Timeout: 10 * time.Second})
	if err != nil {
		t.Fatalf("Failed to create REST client: %v", err)
	}

	// 执行真实的 GET 请求
	ctx := context.Background()
	result := client.Get().
		Resource("service").
		Param("pageNum", "1").
		Param("pageSize", "10").
		Do(ctx)

	// 解析响应
	var serviceList ServiceListResponse
	err = result.Into(&serviceList)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// 打印结果（用于调试）
	t.Logf("Total services: %d", serviceList.Total)
	t.Logf("Page size: %d", serviceList.PageSize)
	t.Logf("Page number: %d", serviceList.PageNum)
	for i, service := range serviceList.List {
		t.Logf("Service %d: ID=%s, Name=%s, Status=%s", i+1, service.ID, service.Name, service.Status)
	}
}

// TestRESTClient_ErrorHandling 测试错误处理
func TestRESTClient_ErrorHandling(t *testing.T) {
	// 创建返回错误的模拟服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := map[string]interface{}{
			"status":      400,
			"message":     "Bad Request",
			"data":        nil,
			"fieldErrors": "Invalid parameters",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// 创建 REST 客户端
	addr := mockServer.Listener.Addr().(*net.TCPAddr)
	client, err := NewRESTClient("http", addr.IP.String(),
		strconv.Itoa(addr.Port), &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Failed to create REST client: %v", err)
	}

	// 执行请求
	ctx := context.Background()
	result := client.Get().
		Resource("service").
		Do(ctx)

	// 尝试解析响应，应该返回错误
	var serviceList ServiceListResponse
	err = result.Into(&serviceList)
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}

	// 验证错误类型和内容
	if apiErr, ok := err.(*Aerror); ok {
		if apiErr.Status != 400 {
			t.Errorf("Expected status 400, got %d", apiErr.Status)
		}
		if apiErr.Message != "Bad Request" {
			t.Errorf("Expected message 'Bad Request', got %s", apiErr.Message)
		}
		if apiErr.FieldErrors != "Invalid parameters" {
			t.Errorf("Expected fieldErrors 'Invalid parameters', got %s", apiErr.FieldErrors)
		}
	} else {
		t.Errorf("Expected *aerror, got %T", err)
	}
}
