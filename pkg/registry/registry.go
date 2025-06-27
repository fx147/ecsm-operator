package registry

type Registry struct {
	store Store
}

func NewRegistry(store Store) *Registry {
	return &Registry{store: store}
}
