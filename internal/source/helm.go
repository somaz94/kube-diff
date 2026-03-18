package source

import (
	"fmt"
	"os/exec"
	"strings"
)

// HelmSource loads resources by running helm template.
type HelmSource struct {
	ChartPath   string
	ReleaseName string
	ValuesFiles []string
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
	for _, vf := range h.ValuesFiles {
		args = append(args, "-f", vf)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("helm template failed: %s: %w", string(output), err)
	}

	return parseYAML(strings.NewReader(string(output)))
}
