# ECSM REST Client 测试

这个目录包含了 ECSM REST 客户端的实现和测试代码。

## 文件说明

- `rest_client.go` - REST 客户端的主要实现
- `request.go` - HTTP 请求构建和执行逻辑
- `response.go` - API 响应处理和错误定义
- `rest_client_test.go` - 单元测试
- `example_test.go` - 使用示例和手动测试

## 运行测试

### 运行所有单元测试
```bash
go test -v ./pkg/ecsm-client/rest/
```

### 运行特定测试
```bash
# 运行模拟服务器测试
go test -v -run TestRESTClient_GetServices ./pkg/ecsm-client/rest/

# 运行错误处理测试
go test -v -run TestRESTClient_ErrorHandling ./pkg/ecsm-client/rest/
```

### 运行真实 API 测试（需要 ECSM 服务器运行）
```bash
# 编辑 rest_client_test.go，取消注释 TestRESTClient_RealAPI 中的 t.Skip() 行
go test -v -run TestRESTClient_RealAPI ./pkg/ecsm-client/rest/

# 或者运行手动测试
# 编辑 example_test.go，取消注释 TestManualRESTClient 中的 t.Skip() 行
go test -v -run TestManualRESTClient ./pkg/ecsm-client/rest/
```

## 使用命令行测试工具

我们提供了一个命令行工具来测试与真实 ECSM API 的连接：

```bash
# 使用默认参数测试
go run ./cmd/test-rest-client/main.go

# 使用自定义参数测试
go run ./cmd/test-rest-client/main.go \
  --host 192.168.31.129 \
  --port 3001 \
  --page 1 \
  --size 10 \
  --verbose

# 查看所有可用参数
go run ./cmd/test-rest-client/main.go --help
```

### 命令行参数说明

- `--host`: ECSM API 服务器主机地址（默认：192.168.31.129）
- `--port`: ECSM API 服务器端口（默认：3001）
- `--protocol`: 协议类型（默认：http）
- `--page`: 页码（默认：1）
- `--size`: 每页大小（默认：10）
- `--timeout`: 请求超时时间（默认：10s）
- `--verbose`: 启用详细日志输出

## 基本使用示例

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "github.com/fx147/ecsm-operator/pkg/ecsm-client/rest"
)

func main() {
    // 创建 REST 客户端
    client, err := rest.NewRESTClient("http", "192.168.31.129", "3001", &http.Client{
        Timeout: 10 * time.Second,
    })
    if err != nil {
        panic(err)
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
        panic(err)
    }
    
    fmt.Printf("Found %d services\n", len(serviceList.List))
}
```

## API 端点

当前支持的 ECSM API 端点：

- `GET /api/v1/service` - 获取服务列表
- `GET /api/v1/service/{id}` - 获取特定服务详情
- 更多端点可以通过相同的方式调用

## 错误处理

客户端会自动处理 ECSM API 的错误响应格式：

```json
{
  "status": 400,
  "message": "Bad Request",
  "fieldErrors": "Invalid parameters",
  "data": null
}
```

错误会被包装为 `aerror` 类型，包含状态码、消息和字段错误信息。

## 日志记录

客户端使用 `klog` 进行日志记录：

- 错误日志：使用 `klog.ErrorS`
- 调试信息：使用 `klog.V(4).InfoS`
- API 错误状态：使用 `klog.V(2).InfoS`

启用详细日志：
```bash
go run ./cmd/test-rest-client/main.go --verbose
# 或者设置环境变量
export KLOG_V=4
```