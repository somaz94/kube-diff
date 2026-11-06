package cli

import (
	"os"
	"path/filepath"
	"testing"
)

const nsKubeconfig = `apiVersion: v1
kind: Config
clusters:
- name: c
  cluster:
    server: https://127.0.0.1:6443
users:
- name: u
  user:
    token: t
contexts:
- name: with-ns
  context:
    cluster: c
    user: u
    namespace: team-a
- name: without-ns
  context:
    cluster: c
    user: u
current-context: with-ns
`

func TestResolveNamespace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(nsKubeconfig), 0o600); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}

	cases := []struct {
		name string
		f    diffFlags
		want string
	}{
		{"flag wins", diffFlags{kubeconfig: path, namespace: "explicit"}, "explicit"},
		{"context namespace", diffFlags{kubeconfig: path}, "team-a"},
		{"context without namespace", diffFlags{kubeconfig: path, kubeContext: "without-ns"}, "default"},
		{"unreadable kubeconfig", diffFlags{kubeconfig: filepath.Join(t.TempDir(), "missing")}, "default"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveNamespace(c.f); got != c.want {
				t.Errorf("resolveNamespace() = %q, want %q", got, c.want)
			}
		})
	}
}
