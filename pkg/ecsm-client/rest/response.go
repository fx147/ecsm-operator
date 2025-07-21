package rest

import (
	"encoding/json"
	"fmt"
)

// aerror 是我们自定义的错误类型，它包含了 ECSM API 返回的详细错误信息。
// 使用小写开头，因为它只在包内使用。
type aerror struct {
	Status      int    `json:"status"`
	Message     string `json:"message"`
	FieldErrors string `json:"fieldErrors"`
}

// Error 方法让 aerror 实现了 Go 的 error 接口。
func (e *aerror) Error() string {
	if e.FieldErrors != "" {
		return fmt.Sprintf("ecsm api error (status %d): %s (field: %s)", e.Status, e.Message, e.FieldErrors)
	}
	return fmt.Sprintf("ecsm api error (status %d): %s", e.Status, e.Message)
}

// response 是用于解码所有 ECSM API 调用的通用响应体结构。
type response struct {
	Status      int             `json:"status"`
	Message     string          `json:"message"`
	Data        json.RawMessage `json:"data"` // 使用 json.RawMessage 来延迟解码 data 部分
	FieldErrors string          `json:"fieldErrors"`
}
