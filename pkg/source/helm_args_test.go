package source

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// fakeHelm puts a helm script first on PATH. It prints a warning to stderr and
// a ConfigMap whose data.args holds the arguments it was called with, or exits
// 1 when fail is set.
func fakeHelm(t *testing.T, fail bool) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake helm is a shell script")
	}
	script := `#!/bin/sh
echo "WARNING: this chart is deprecated" >&2
cat <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: args
data:
  args: "$*"
EOF
`
	if fail {
		script = "#!/bin/sh\necho 'Error: chart not found' >&2\nexit 1\n"
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "helm"), []byte(script), 0o755); err != nil {
		t.Fatalf("write fake helm: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func helmArgs(t *testing.T, h *HelmSource) string {
	t.Helper()
	resources, err := h.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Load() returned %d resources, want 1 (stderr must not be parsed)", len(resources))
	}
	args, _, err := unstructured.NestedString(resources[0].Object.Object, "data", "args")
	if err != nil {
		t.Fatalf("read data.args: %v", err)
	}
	return args
}

func TestHelmSourcePassesNamespace(t *testing.T) {
	fakeHelm(t, false)
	h := NewHelmSource("./chart", "rel", []string{"v.yaml"})
	h.Namespace = "team-a"

	got := helmArgs(t, h)
	if want := "template rel ./chart --namespace team-a -f v.yaml"; got != want {
		t.Errorf("helm args = %q, want %q", got, want)
	}
}

func TestHelmSourceOmitsEmptyNamespace(t *testing.T) {
	fakeHelm(t, false)

	got := helmArgs(t, NewHelmSource("./chart", "rel", nil))
	if strings.Contains(got, "--namespace") {
		t.Errorf("helm args = %q, want no --namespace", got)
	}
}

func TestHelmSourceErrorIncludesStderr(t *testing.T) {
	fakeHelm(t, true)

	_, err := NewHelmSource("./chart", "rel", nil).Load()
	if err == nil || !strings.Contains(err.Error(), "Error: chart not found") {
		t.Fatalf("Load() error = %v, want it to include helm's stderr", err)
	}
}
