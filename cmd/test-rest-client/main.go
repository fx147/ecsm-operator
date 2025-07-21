package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/rest"
	"k8s.io/klog/v2"
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

func main() {
	// 初始化 klog
	klog.InitFlags(nil)
	
	// 定义命令行参数
	var (
		host     = flag.String("host", "192.168.31.129", "ECSM API server host")
		port     = flag.String("port", "3001", "ECSM API server port")
		protocol = flag.String("protocol", "http", "Protocol (http or https)")
		pageNum  = flag.String("page", "1", "Page number")
		pageSize = flag.String("size", "10", "Page size")
		timeout  = flag.Duration("timeout", 10*time.Second, "Request timeout")
		verbose  = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()
	
	// 设置日志级别
	if *verbose {
		flag.Set("v", "4")
	}

	fmt.Printf("Testing ECSM REST Client\n")
	fmt.Printf("========================\n")
	fmt.Printf("Server: %s://%s:%s\n", *protocol, *host, *port)
	fmt.Printf("Page: %s, Size: %s\n", *pageNum, *pageSize)
	fmt.Printf("Timeout: %v\n\n", *timeout)

	// 创建 REST 客户端
	client, err := rest.NewRESTClient(*protocol, *host, *port, &http.Client{
		Timeout: *timeout,
	})
	if err != nil {
		fmt.Printf("❌ Failed to create REST client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ REST client created successfully\n\n")

	// 执行 GET 请求获取服务列表
	fmt.Printf("🔄 Fetching service list...\n")
	ctx := context.Background()
	result := client.Get().
		Resource("service").
		Param("pageNum", *pageNum).
		Param("pageSize", *pageSize).
		Do(ctx)

	// 解析响应
	var serviceList ServiceListResponse
	err = result.Into(&serviceList)
	if err != nil {
		fmt.Printf("❌ Failed to fetch services: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully fetched service list\n\n")

	// 显示结果摘要
	fmt.Printf("📊 Results Summary\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total services: %d\n", serviceList.Total)
	fmt.Printf("Page size: %d\n", serviceList.PageSize)
	fmt.Printf("Page number: %d\n", serviceList.PageNum)
	fmt.Printf("Services in this page: %d\n\n", len(serviceList.List))

	// 显示服务详细信息
	if len(serviceList.List) > 0 {
		fmt.Printf("📋 Service Details\n")
		fmt.Printf("==================\n")
		for i, service := range serviceList.List {
			fmt.Printf("Service %d:\n", i+1)
			fmt.Printf("  🆔 ID: %s\n", service.ID)
			fmt.Printf("  📛 Name: %s\n", service.Name)
			fmt.Printf("  📊 Status: %s\n", service.Status)
			fmt.Printf("  📅 Created: %s\n", service.CreatedTime)
			fmt.Printf("  🔄 Updated: %s\n", service.UpdatedTime)
			fmt.Printf("  📋 Policy: %s\n", service.Policy)
			fmt.Printf("  🔢 Factor: %d\n", service.Factor)
			fmt.Println()
		}
	} else {
		fmt.Printf("ℹ️  No services found\n")
	}

	fmt.Printf("🎉 Test completed successfully!\n")
}