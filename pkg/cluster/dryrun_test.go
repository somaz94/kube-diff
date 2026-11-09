package cluster

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clienttesting "k8s.io/client-go/testing"
)

func TestDryRunApplySendsForcedDryRunApply(t *testing.T) {
	f := listFetcher()
	var got clienttesting.PatchActionImpl
	f.client.(interface {
		PrependReactor(verb, resource string, fn clienttesting.ReactionFunc)
	}).PrependReactor("patch", "configmaps", func(a clienttesting.Action) (bool, runtime.Object, error) {
		got = a.(clienttesting.PatchActionImpl)
		return true, labeled("ConfigMap", "default", "cm", map[string]interface{}{"defaulted": "yes"}), nil
	})

	out, err := f.DryRunApply(context.Background(), labeled("ConfigMap", "default", "cm", nil), "my-manager")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.GetLabels()["defaulted"] != "yes" {
		t.Errorf("expected the server's object back, got %v", out.Object)
	}
	if got.GetPatchType() != types.ApplyPatchType || got.GetNamespace() != "default" || got.GetName() != "cm" {
		t.Errorf("unexpected request: type=%s ns=%s name=%s", got.GetPatchType(), got.GetNamespace(), got.GetName())
	}
	opts := got.PatchOptions
	if len(opts.DryRun) != 1 || opts.DryRun[0] != metav1.DryRunAll {
		t.Errorf("expected dryRun=All, got %v", opts.DryRun)
	}
	if opts.FieldManager != "my-manager" || opts.Force == nil || !*opts.Force {
		t.Errorf("expected forced apply as my-manager, got manager=%q force=%v", opts.FieldManager, opts.Force)
	}
}

func TestDryRunApplyClusterScoped(t *testing.T) {
	f := listFetcher()
	var ns = "unset"
	f.client.(interface {
		PrependReactor(verb, resource string, fn clienttesting.ReactionFunc)
	}).PrependReactor("patch", "namespaces", func(a clienttesting.Action) (bool, runtime.Object, error) {
		ns = a.GetNamespace()
		return true, labeled("Namespace", "", "team-a", nil), nil
	})
	if _, err := f.DryRunApply(context.Background(), labeled("Namespace", "", "team-a", nil), "m"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ns != "" {
		t.Errorf("a cluster-scoped apply should carry no namespace, got %q", ns)
	}
}

func TestDryRunApplyBadAPIVersion(t *testing.T) {
	f := listFetcher()
	obj := labeled("ConfigMap", "default", "cm", nil)
	obj.SetAPIVersion("a/b/c")
	if _, err := f.DryRunApply(context.Background(), obj, "m"); err == nil {
		t.Fatal("expected an error for an unparsable apiVersion")
	}
}
