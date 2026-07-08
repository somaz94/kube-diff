package source

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// FileSource loads resources from plain YAML files.
type FileSource struct {
	Path string
}

func NewFileSource(path string) *FileSource {
	return &FileSource{Path: path}
}

func (f *FileSource) Load() ([]Resource, error) {
	info, err := os.Stat(f.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot access %s: %w", f.Path, err)
	}

	if info.IsDir() {
		return f.loadDirectory()
	}
	return f.loadFile(f.Path)
}

func (f *FileSource) loadDirectory() ([]Resource, error) {
	var resources []Resource

	err := filepath.Walk(f.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		res, err := f.loadFile(path)
		if err != nil {
			return fmt.Errorf("error loading %s: %w", path, err)
		}
		resources = append(resources, res...)
		return nil
	})

	return resources, err
}

func (f *FileSource) loadFile(path string) ([]Resource, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseYAML(file)
}

func parseYAML(reader io.Reader) ([]Resource, error) {
	var resources []Resource

	yamlReader := yaml.NewYAMLReader(bufio.NewReader(reader))
	for {
		data, err := yamlReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading YAML: %w", err)
		}

		if len(strings.TrimSpace(string(data))) == 0 {
			continue
		}

		obj := &unstructured.Unstructured{}
		if err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), len(data)).Decode(obj); err != nil {
			// Skip documents that can't be decoded as K8s resources (e.g., missing Kind)
			continue
		}

		if obj.GetKind() == "" {
			continue
		}

		resources = append(resources, Resource{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			Name:       obj.GetName(),
			Namespace:  obj.GetNamespace(),
			Object:     obj,
		})
	}

	return resources, nil
}
