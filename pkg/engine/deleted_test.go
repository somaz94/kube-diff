package engine

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/somaz94/kube-diff/internal/testutil"
	"github.com/somaz94/kube-diff/pkg/diff"
	"github.com/somaz94/kube-diff/pkg/source"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// fakeLister returns items per "apiVersion/kind/namespace" and records the selectors it saw.
type fakeLister struct {
	items     map[string][]unstructured.Unstructured
	errs      map[string]error
	selectors []string
}

func (l *fakeLister) List(_ context.Context, apiVersion, kind, namespace, selector string) ([]unstructured.Unstructured, error) {
	l.selectors = append(l.selectors, selector)
	key := apiVersion + "/" + kind + "/" + namespace
	if err, ok := l.errs[key]; ok {
		return nil, err
	}
	return l.items[key], nil
}

func cm(namespace, name string) unstructured.Unstructured {
	return *testutil.NewTestObj("v1", "ConfigMap", name, namespace, nil)
}

func TestFindDeletedReportsMissingObjects(t *testing.T) {
	owned := cm("default", "generated")
	owned.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion: "apps/v1", Kind: "Deployment", Name: "web", UID: "u1", Controller: boolPtr(true),
	}})
	lister := &fakeLister{items: map[string][]unstructured.Unstructured{
		"v1/ConfigMap/default": {cm("default", "kept"), cm("default", "gone"), owned},
	}}
	local := testutil.NewTestObj("v1", "ConfigMap", "kept", "default", nil)

	results, err := FindDeleted(context.Background(), lister, []source.Resource{resource("kept", local)}, "app=web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Name != "gone" || results[0].Status != diff.StatusDeleted {
		t.Fatalf("expected only gone as deleted, got %+v", results)
	}
	if results[0].Namespace != "default" || results[0].Kind != "ConfigMap" {
		t.Errorf("unexpected identity: %+v", results[0])
	}
	if strings.Join(lister.selectors, ",") != "app=web" {
		t.Errorf("expected one list call with the selector, got %v", lister.selectors)
	}
}

func TestFindDeletedListsEachScopeOnce(t *testing.T) {
	lister := &fakeLister{items: map[string][]unstructured.Unstructured{}}
	a := testutil.NewTestObj("v1", "ConfigMap", "a", "default", nil)
	b := testutil.NewTestObj("v1", "ConfigMap", "b", "default", nil)
	c := testutil.NewTestObj("v1", "ConfigMap", "c", "prod", nil)

	if _, err := FindDeleted(context.Background(), lister, []source.Resource{resource("a", a), resource("b", b), resource("c", c)}, "x=y"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lister.selectors) != 2 {
		t.Errorf("expected two list calls (default, prod), got %d", len(lister.selectors))
	}
}

func TestFindDeletedSkipsUnservedKind(t *testing.T) {
	lister := &fakeLister{errs: map[string]error{
		"example.com/v1/Widget/default": apierrors.NewNotFound(schema.GroupResource{Resource: "widgets"}, ""),
	}}
	w := testutil.NewTestObj("example.com/v1", "Widget", "w", "default", nil)

	results, err := FindDeleted(context.Background(), lister, []source.Resource{resource("w", w)}, "x=y")
	if err != nil || len(results) != 0 {
		t.Fatalf("expected no results and no error, got %v, %v", results, err)
	}
}

func TestFindDeletedPropagatesListErrors(t *testing.T) {
	lister := &fakeLister{errs: map[string]error{
		"v1/ConfigMap/default": apierrors.NewForbidden(schema.GroupResource{Resource: "configmaps"}, "", errors.New("no list")),
	}}
	a := testutil.NewTestObj("v1", "ConfigMap", "a", "default", nil)

	_, err := FindDeleted(context.Background(), lister, []source.Resource{resource("a", a)}, "x=y")
	if err == nil || !apierrors.IsForbidden(err) || !strings.Contains(err.Error(), `namespace "default"`) {
		t.Fatalf("expected a wrapped Forbidden error naming the namespace, got %v", err)
	}
}

func boolPtr(b bool) *bool { return &b }
