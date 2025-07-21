package clientset

type RecordGetter interface {
	Records() RecordInterface
}

type RecordInterface interface{}
