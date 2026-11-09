package diff

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/somaz94/kube-diff/internal/testutil"
)

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

func secret(fields map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"type": "Opaque", "data": fields}
}

func TestCompareSecretRedactsChangedValue(t *testing.T) {
	local := testutil.NewTestObj("v1", "Secret", "db", "default", secret(map[string]interface{}{
		"password": b64("new-pass"), "user": b64("admin"),
	}))
	cluster := testutil.NewTestObj("v1", "Secret", "db", "default", secret(map[string]interface{}{
		"password": b64("old-pass"), "user": b64("admin"),
	}))

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusChanged {
		t.Fatalf("expected StatusChanged, got %s", result.Status)
	}
	for _, leak := range []string{b64("new-pass"), b64("old-pass"), "new-pass", "old-pass"} {
		if strings.Contains(result.Diff, leak) {
			t.Errorf("diff leaks %q:\n%s", leak, result.Diff)
		}
	}
	if !strings.Contains(result.Diff, "+    password: "+RedactedChanged) {
		t.Errorf("diff should mark the changed key, got:\n%s", result.Diff)
	}
	if strings.Contains(result.Diff, "+    user:") {
		t.Errorf("an unchanged key should not appear as changed:\n%s", result.Diff)
	}
}

func TestCompareSecretUnchanged(t *testing.T) {
	fields := map[string]interface{}{"token": b64("same")}
	local := testutil.NewTestObj("v1", "Secret", "s", "default", secret(fields))
	cluster := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{"token": b64("same")}))

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusUnchanged {
		t.Errorf("expected StatusUnchanged, got %s:\n%s", result.Status, result.Diff)
	}
}

func TestCompareSecretStringDataMatchesStoredData(t *testing.T) {
	local := testutil.NewTestObj("v1", "Secret", "s", "default", map[string]interface{}{
		"type":       "Opaque",
		"stringData": map[string]interface{}{"token": "same"},
	})
	cluster := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{"token": b64("same")}))

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusUnchanged {
		t.Errorf("stringData equal to the stored data should be unchanged, got %s:\n%s", result.Status, result.Diff)
	}
}

func TestCompareSecretStringDataChangeIsRedacted(t *testing.T) {
	local := testutil.NewTestObj("v1", "Secret", "s", "default", map[string]interface{}{
		"type":       "Opaque",
		"data":       map[string]interface{}{"keep": b64("k")},
		"stringData": map[string]interface{}{"token": "plain-secret"},
	})
	cluster := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{
		"keep": b64("k"), "token": b64("other"),
	}))

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusChanged {
		t.Fatalf("expected StatusChanged, got %s", result.Status)
	}
	for _, leak := range []string{"plain-secret", b64("plain-secret"), b64("other"), "stringData"} {
		if strings.Contains(result.Diff, leak) {
			t.Errorf("diff leaks %q:\n%s", leak, result.Diff)
		}
	}
}

func TestCompareSecretAddedAndRemovedKeys(t *testing.T) {
	local := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{"added": b64("a")}))
	cluster := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{"removed": b64("r")}))

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Diff, "+    added: "+Redacted) || !strings.Contains(result.Diff, "-    removed: "+Redacted) {
		t.Errorf("added and removed keys should be named with redacted values, got:\n%s", result.Diff)
	}
	if strings.Contains(result.Diff, b64("a")) || strings.Contains(result.Diff, b64("r")) {
		t.Errorf("diff leaks a value:\n%s", result.Diff)
	}
}

func TestCompareSecretShowSecrets(t *testing.T) {
	local := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{"token": b64("new")}))
	cluster := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{"token": b64("old")}))

	opts := DefaultCompareOptions()
	opts.ShowSecrets = true
	result, err := Compare(local, cluster, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Diff, b64("new")) || !strings.Contains(result.Diff, b64("old")) {
		t.Errorf("ShowSecrets should print both values, got:\n%s", result.Diff)
	}
}

func TestCompareSecretLastAppliedIsRedacted(t *testing.T) {
	local := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{"token": b64("new")}))
	cluster := testutil.NewTestObj("v1", "Secret", "s", "default", secret(map[string]interface{}{"token": b64("live")}))
	cluster.SetAnnotations(map[string]string{
		"kubectl.kubernetes.io/last-applied-configuration": `{"apiVersion":"v1","kind":"Secret","metadata":{"name":"s","namespace":"default"},"type":"Opaque","stringData":{"token":"applied-plain"}}`,
	})

	opts := DefaultCompareOptions()
	opts.Strategy = StrategyLastApplied
	result, err := Compare(local, cluster, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, leak := range []string{"applied-plain", b64("applied-plain"), b64("new"), b64("live")} {
		if strings.Contains(result.Diff, leak) {
			t.Errorf("diff leaks %q:\n%s", leak, result.Diff)
		}
	}
}

func TestCompareConfigMapNotRedacted(t *testing.T) {
	local := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{"data": map[string]interface{}{"k": "new"}})
	cluster := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{"data": map[string]interface{}{"k": "old"}})

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Diff, "new") || strings.Contains(result.Diff, Redacted) {
		t.Errorf("ConfigMap data should be shown as is, got:\n%s", result.Diff)
	}
}
