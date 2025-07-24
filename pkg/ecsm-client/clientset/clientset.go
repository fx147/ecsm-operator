package clientset

import "github.com/fx147/ecsm-operator/pkg/ecsm-client/rest"

type Interface interface {
	RESTClient() rest.RESTClient
	ServiceGetter
	RecordGetter
	ContainerGetter
	NodeGetter
}

type Clientset struct {
	restClient rest.RESTClient
}

// NewClientset 创建一个新的 Clientset 实例，用于与 ECSM API 交互
func NewClientset(protocol, host, port string) (*Clientset, error) {
	// 创建 REST 客户端
	restClient, err := rest.NewRESTClient(protocol, host, port, nil)
	if err != nil {
		return nil, err
	}

	// 创建并返回 Clientset
	return &Clientset{
		restClient: *restClient,
	}, nil
}

// RESTClient 返回底层的 REST 客户端
func (c *Clientset) RESTClient() rest.RESTClient {
	return c.restClient
}

// Services 返回 ServiceInterface，用于操作 Service 资源
func (c *Clientset) Services() ServiceInterface {
	return newServices(&c.restClient)
}

// Records 返回 RecordInterface，用于操作 Record 资源
func (c *Clientset) Records() RecordInterface {
	return nil // 暂未实现
}

// Containers 返回 ContainerInterface，用于操作 Container 资源
func (c *Clientset) Containers() ContainerInterface {
	return newContainers(&c.restClient)
}

func (c *Clientset) Nodes() NodeInterface {
	return newNodes(&c.restClient)
}

func (c *Clientset) Images() ImageInterface {
	return newImages(&c.restClient)
}
