// file: pkg/ecsm_client/rest/request.go

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"k8s.io/klog/v2"
)

// Request 允许以链式方式构建请求。
type Request struct {
	c          *RESTClient
	verb       string
	resource   string
	resourceID string
	body       interface{}
	err        error
	params     url.Values
}

func NewRequest(c *RESTClient) *Request {
	return &Request{
		c: c,
	}
}

// Verb 指定 HTTP 方法 (e.g., "GET", "POST")。
func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

// Resource 指定要操作的资源 (e.g., "services")。
func (r *Request) Resource(resource string) *Request {
	if r.err != nil {
		return r
	}
	r.resource = resource
	return r
}

// Name 指定要操作的资源的具体名称/ID。
func (r *Request) Name(name string) *Request {
	if r.err != nil {
		return r
	}
	if len(name) == 0 {
		r.err = fmt.Errorf("resource name may not be empty")
		return r
	}
	if len(r.resourceID) != 0 {
		r.err = fmt.Errorf("resource name already set to %q, cannot change to %q", r.resourceID, name)
		return r
	}
	r.resourceID = name
	return r
}

// Body 设置请求体。传入的 obj 会被序列化为 JSON。
func (r *Request) Body(obj interface{}) *Request {
	if r.err != nil {
		return r
	}
	r.body = obj
	return r
}

// Param 向请求添加一个 URL Query 参数。
func (r *Request) Param(key, value string) *Request {
	if r.err != nil {
		return r
	}
	if r.params == nil {
		r.params = make(url.Values)
	}
	r.params.Add(key, value)
	return r
}

// Do 执行请求并返回一个 Result 对象。
func (r *Request) Do(ctx context.Context) *Result {
	if r.err != nil {
		return &Result{err: r.err}
	}

	// ---- 完整的请求构建和执行逻辑 ----
	// 1. 构建 URL
	p := path.Join(r.c.apiPath, r.c.apiVersion, r.resource)
	// p := path.Join(r.c.apiVersion, r.resource)
	if r.resourceID != "" {
		p = path.Join(p, r.resourceID)
	}
	fullURL := r.c.baseURL.ResolveReference(&url.URL{Path: p})

	if len(r.params) > 0 {
		fullURL.RawQuery = r.params.Encode()
	}

	// 2. 序列化 Body
	var bodyReader io.Reader
	if r.body != nil {
		data, err := json.Marshal(r.body)
		if err != nil {
			r.err = fmt.Errorf("failed to marshal body: %w", err)
			return &Result{err: r.err}
		}
		bodyReader = bytes.NewBuffer(data)
	}

	// 3. 创建 HTTP Request
	req, err := http.NewRequestWithContext(ctx, r.verb, fullURL.String(), bodyReader)
	if err != nil {
		r.err = fmt.Errorf("failed to create request: %w", err)
		return &Result{err: r.err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 4. 执行请求
	klog.V(4).InfoS("Executing request", "method", req.Method, "url", req.URL)
	resp, err := r.c.httpClient.Do(req)
	if err != nil {
		r.err = fmt.Errorf("request failed: %w", err)
		return &Result{err: r.err}
	}

	return &Result{
		body:       resp.Body,
		statusCode: resp.StatusCode,
		err:        nil,
	}
}

// Result 封装了请求的结果。
type Result struct {
	body       io.ReadCloser
	statusCode int
	err        error
}

// transformAndGetRawData 是一个新的辅助方法。
// 它解码通用的响应信封，检查 API 错误，如果成功，则返回原始的 data 字段。
func (r *Result) transformAndGetRawData() (json.RawMessage, error) {
	// 先调用 Raw() 获取原始 body
	bodyBytes, err := r.Raw()
	if err != nil {
		return nil, err
	}

	// 如果 body 为空，直接返回成功
	if len(bodyBytes) == 0 {
		return nil, nil
	}

	// 解码到通用的 response 结构体
	var apiResp Response
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode generic response: %w (raw response: %q)", err, string(bodyBytes))
	}

	// 检查 API 级别的错误
	if apiResp.Status != 200 {
		return nil, &Aerror{
			Status:      apiResp.Status,
			Message:     apiResp.Message,
			FieldErrors: apiResp.FieldErrors,
		}
	}

	// 返回原始的 data 字段以供进一步处理
	return apiResp.Data, nil
}

// Into 解码响应体到传入的 obj 对象中。
// 我们让它内部调用 transformAndGetRawData 来复用逻辑。
func (r *Result) Into(obj interface{}) error {
	rawData, err := r.transformAndGetRawData()
	if err != nil {
		return err
	}

	// 如果请求成功，但没有 data 或者调用者不关心，则直接返回
	if obj == nil || len(rawData) == 0 || string(rawData) == "null" {
		return nil
	}

	// 解码 data 部分
	if err := json.Unmarshal(rawData, obj); err != nil {
		return fmt.Errorf("failed to unmarshal data into object: %w", err)
	}

	return nil
}

// Raw 读取并返回原始的响应体 []byte。
// 注意：这个操作会消耗掉响应体，不能与 Into() 同时使用。
func (r *Result) Raw() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	defer r.body.Close()
	return io.ReadAll(r.body)
}
