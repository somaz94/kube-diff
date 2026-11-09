package engine

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/somaz94/kube-diff/internal/testutil"
	"github.com/somaz94/kube-diff/pkg/diff"
	"github.com/somaz94/kube-diff/pkg/source"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// applyingFetcher answers Get from objs and DryRunApply by adding the server's
// defaults to the manifest, recording the field manager it was called with.
type applyingFetcher struct {
	fakeFetcher
	defaults map[string]interface{}
	err      error
	managers []string
}

func (a *applyingFetcher) DryRunApply(_ context.Context, obj *unstructured.Unstructured, fieldManager string) (*unstructured.Unstructured, error) {
	a.managers = append(a.managers, fieldManager)
	if a.err != nil {
		return nil, a.err
	}
	out := obj.DeepCopy()
	data, _ := out.Object["data"].(map[string]interface{})
	for k, v := range a.defaults {
		data[k] = v
	}
	return out, nil
}

func defaultedCase() (*applyingFetcher, []source.Resource) {
	live := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"k": "v", "injected": "by-webhook"},
	})
	local := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"k": "v"},
	})
	f := &applyingFetcher{
		fakeFetcher: fakeFetcher{objs: map[string]*unstructured.Unstructured{"cm": live}},
		defaults:    map[string]interface{}{"injected": "by-webhook"},
	}
	return f, []source.Resource{resource("cm", local)}
}

func TestCompareServerSideIncludesServerDefaults(t *testing.T) {
	f, resources := defaultedCase()

	live, err := Compare(context.Background(), f, resources, diff.DefaultCompareOptions())
	if err != nil || live[0].Status != diff.StatusChanged {
		t.Fatalf("without --server-side the injected key should show as changed, got %v, %v", live, err)
	}

	opts := diff.DefaultCompareOptions()
	opts.ServerSide, opts.FieldManager = true, "helm"
	results, err := Compare(context.Background(), f, resources, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != diff.StatusUnchanged {
		t.Errorf("expected unchanged with server defaults applied, got %s:\n%s", results[0].Status, results[0].Diff)
	}
	if strings.Join(f.managers, ",") != "helm" {
		t.Errorf("expected one dry run as helm, got %v", f.managers)
	}
}

func TestCompareServerSideSkipsNewResources(t *testing.T) {
	f := &applyingFetcher{}
	local := testutil.NewTestObj("v1", "ConfigMap", "fresh", "default", nil)
	opts := diff.DefaultCompareOptions()
	opts.ServerSide = true

	results, err := Compare(context.Background(), f, []source.Resource{resource("fresh", local)}, opts)
	if err != nil || results[0].Status != diff.StatusNew || len(f.managers) != 0 {
		t.Fatalf("expected NEW without a dry run, got %v, %v, managers=%v", results, err, f.managers)
	}
}

func TestCompareServerSideWrapsDryRunErrors(t *testing.T) {
	f, resources := defaultedCase()
	f.err = errors.New("webhook does not support dry run")
	opts := diff.DefaultCompareOptions()
	opts.ServerSide = true

	_, err := Compare(context.Background(), f, resources, opts)
	if err == nil || !strings.Contains(err.Error(), "dry-run apply v1 ConfigMap default/cm") {
		t.Fatalf("expected a wrapped error naming the resource, got %v", err)
	}
}

func TestCompareServerSideRejectsLastApplied(t *testing.T) {
	f, resources := defaultedCase()
	opts := diff.DefaultCompareOptions()
	opts.ServerSide, opts.Strategy = true, diff.StrategyLastApplied
	if _, err := Compare(context.Background(), f, resources, opts); err == nil {
		t.Fatal("expected server-side plus last-applied to be rejected")
	}
}

func TestCompareServerSideNeedsApplier(t *testing.T) {
	opts := diff.DefaultCompareOptions()
	opts.ServerSide = true
	if _, err := Compare(context.Background(), &fakeFetcher{}, nil, opts); err == nil {
		t.Fatal("expected an error for a fetcher that cannot dry-run")
	}
}
