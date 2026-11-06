package source

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// HelmSource loads resources by running helm template.
type HelmSource struct {
	ChartPath   string
	ReleaseName string
	ValuesFiles []string
	// Namespace is passed to helm template as --namespace, so templates that
	// read .Release.Namespace render the namespace the release is installed in.
	Namespace string
}

func NewHelmSource(chartPath, releaseName string, valuesFiles []string) *HelmSource {
	return &HelmSource{
		ChartPath:   chartPath,
		ReleaseName: releaseName,
		ValuesFiles: valuesFiles,
	}
}

func (h *HelmSource) Load() ([]Resource, error) {
	args := []string{"template", h.ReleaseName, h.ChartPath}
	if h.Namespace != "" {
		args = append(args, "--namespace", h.Namespace)
	}
	for _, vf := range h.ValuesFiles {
		args = append(args, "-f", vf)
	}

	// Stderr is kept apart: helm prints warnings there, and mixed into stdout
	// they would be parsed as manifest text.
	var stderr bytes.Buffer
	cmd := exec.Command("helm", args...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("helm template failed: %s: %w", strings.TrimSpace(stderr.String()), err)
	}

	return parseYAML(bytes.NewReader(output))
}
