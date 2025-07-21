package clientset

type ContainerGetter interface {
	Containers() ContainerInterface
}

type ContainerInterface interface {
}
