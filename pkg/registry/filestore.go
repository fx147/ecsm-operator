// file: pkg/registry/filestore.go

package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/fx147/ecsm-operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FileStore 实现了 Store 接口，使用本地文件系统作为后端。
type FileStore struct {
	basePath string
	scheme   *runtime.Scheme
}

var _ Store = &FileStore{}

func NewFileStore(basePath string, scheme *runtime.Scheme) (*FileStore, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path for filestore: %w", err)
	}
	return &FileStore{basePath: basePath, scheme: scheme}, nil
}

func (fs *FileStore) getPathForObject(obj runtime.Object) (string, error) {
	gvk, err := util.GetGVK(obj, fs.scheme)
	if err != nil {
		return "", err
	}

	// 这里list应该不会调用
	meta, err := util.GetObjectMeta(obj)
	if err != nil {
		return "", err
	}

	kindPlural := strings.ToLower(gvk.Kind) + "s"
	return filepath.Join(fs.basePath, gvk.Group, gvk.Version, kindPlural, meta.Namespace, meta.Name+".json"), nil
}

func (fs *FileStore) getDirForKind(namespace string, obj runtime.Object) (string, error) {
	gvk, err := util.GetGVK(obj, fs.scheme)
	if err != nil {
		return "", err
	}

	itemKind := strings.TrimSuffix(gvk.Kind, "List")
	kindPlural := strings.ToLower(itemKind) + "s"
	return filepath.Join(fs.basePath, gvk.Group, gvk.Version, kindPlural, namespace), nil
}

// --- 接口实现 (代码现在更健壮) ---

func (fs *FileStore) Create(obj runtime.Object) error {
	path, err := fs.getPathForObject(obj)
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		meta, _ := util.GetObjectMeta(obj)
		gvk, _ := util.GetGVK(obj, fs.scheme)
		gr := schema.GroupResource{Group: gvk.Group, Resource: strings.ToLower(gvk.Kind) + "s"}
		return errors.NewAlreadyExists(gr, meta.Name)
	}

	dir := filepath.Dir(path)
	if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create directory for object: %w", mkdirErr)
	}

	data, marshalErr := json.MarshalIndent(obj, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal object to json: %w", marshalErr)
	}

	return os.WriteFile(path, data, 0644)
}

func (fs *FileStore) Update(obj runtime.Object) error {
	path, err := fs.getPathForObject(obj)
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		meta, _ := util.GetObjectMeta(obj)
		gvk, _ := util.GetGVK(obj, fs.scheme)
		gr := schema.GroupResource{Group: gvk.Group, Resource: strings.ToLower(gvk.Kind) + "s"}
		return errors.NewNotFound(gr, meta.Name)
	}

	data, marshalErr := json.MarshalIndent(obj, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal object to json: %w", marshalErr)
	}

	return os.WriteFile(path, data, 0644)
}

func (fs *FileStore) Get(namespace, name string, objInto runtime.Object) error {
	dir, err := fs.getDirForKind(namespace, objInto)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			gvk, _ := util.GetGVK(objInto, fs.scheme)
			gr := schema.GroupResource{Group: gvk.Group, Resource: strings.ToLower(gvk.Kind) + "s"}
			return errors.NewNotFound(gr, name)
		}
		return fmt.Errorf("failed to read object file: %w", err)
	}
	return json.Unmarshal(data, objInto)
}

func (fs *FileStore) List(namespace string, listInto runtime.Object) error {
	dirPath, err := fs.getDirForKind(namespace, listInto)
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(dirPath); os.IsNotExist(statErr) {
		return nil
	}

	listValue := reflect.ValueOf(listInto).Elem()
	itemsField := listValue.FieldByName("Items")
	itemType := itemsField.Type().Elem()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			path := filepath.Join(dirPath, entry.Name())
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to read file %s: %v\n", path, readErr)
				continue
			}

			newItem := reflect.New(itemType).Interface().(runtime.Object)
			if umErr := json.Unmarshal(data, newItem); umErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to unmarshal file %s: %v\n", path, umErr)
				continue
			}
			itemsField.Set(reflect.Append(itemsField, reflect.ValueOf(newItem).Elem()))
		}
	}
	return nil
}

func (fs *FileStore) Delete(namespace, name string, objToDelete runtime.Object) error {
	dir, err := fs.getDirForKind(namespace, objToDelete)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name+".json")

	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete object file: %w", err)
	}

	return nil
}
