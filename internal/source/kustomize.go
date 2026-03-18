package source

import (
	"fmt"
	"os/exec"
	"strings"
)

// KustomizeSource loads resources by running kustomize build.
type KustomizeSource struct {
	Path string
}

func NewKustomizeSource(path string) *KustomizeSource {
	return &KustomizeSource{Path: path}
}

func (k *KustomizeSource) Load() ([]Resource, error) {
	cmd := exec.Command("kustomize", "build", k.Path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: try kubectl kustomize
		cmd = exec.Command("kubectl", "kustomize", k.Path)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("kustomize build failed: %s: %w", string(output), err)
		}
	}

	return parseYAML(strings.NewReader(string(output)))
}
