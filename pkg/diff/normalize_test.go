package diff

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNormalizeRemovesStatus(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "test",
			},
			"status": map[string]interface{}{
				"loadBalancer": map[string]interface{}{},
			},
		},
	}

	result := Normalize(obj)

	if _, ok := result.Object["status"]; ok {
		t.Error("expected status to be removed")
	}
}

func TestNormalizeRemovesMetadataFields(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":              "test",
				"uid":               "abc-123",
				"resourceVersion":   "999",
				"creationTimestamp": "2024-01-01T00:00:00Z",
				"generation":        int64(5),
				"selfLink":          "/api/v1/configmaps/test",
				"managedFields":     []interface{}{},
			},
		},
	}

	result := Normalize(obj)
	metadata := result.Object["metadata"].(map[string]interface{})

	removedFields := []string{"uid", "resourceVersion", "creationTimestamp", "generation", "selfLink", "managedFields"}
	for _, field := range removedFields {
		if _, ok := metadata[field]; ok {
			t.Errorf("expected %s to be removed from metadata", field)
		}
	}

	if metadata["name"] != "test" {
		t.Error("expected name to be preserved")
	}
}

func TestNormalizeRemovesKubectlAnnotation(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "test",
				"annotations": map[string]interface{}{
					"kubectl.kubernetes.io/last-applied-configuration": "{}",
					"my-annotation": "keep-me",
				},
			},
		},
	}

	result := Normalize(obj)
	metadata := result.Object["metadata"].(map[string]interface{})
	annotations := metadata["annotations"].(map[string]interface{})

	if _, ok := annotations["kubectl.kubernetes.io/last-applied-configuration"]; ok {
		t.Error("expected kubectl annotation to be removed")
	}
	if annotations["my-annotation"] != "keep-me" {
		t.Error("expected custom annotation to be preserved")
	}
}

func TestNormalizeRemovesDeploymentRevisionAnnotation(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "test",
				"annotations": map[string]interface{}{
					"deployment.kubernetes.io/revision": "3",
				},
			},
		},
	}

	result := Normalize(obj)
	metadata := result.Object["metadata"].(map[string]interface{})

	// annotations should be removed entirely since it's empty
	if _, ok := metadata["annotations"]; ok {
		t.Error("expected empty annotations map to be removed")
	}
}

func TestNormalizeNilInput(t *testing.T) {
	result := Normalize(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestNormalizeDeploymentDefaults(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "app",
			},
			"spec": map[string]interface{}{
				"replicas":                int64(2),
				"progressDeadlineSeconds": int64(600),
				"revisionHistoryLimit":    int64(10),
				"strategy": map[string]interface{}{
					"type": "RollingUpdate",
					"rollingUpdate": map[string]interface{}{
						"maxSurge":       "25%",
						"maxUnavailable": "25%",
					},
				},
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":                     "app",
								"image":                    "nginx:1.25",
								"imagePullPolicy":          "IfNotPresent",
								"terminationMessagePath":   "/dev/termination-log",
								"terminationMessagePolicy": "File",
								"ports": []interface{}{
									map[string]interface{}{
										"containerPort": int64(80),
										"protocol":      "TCP",
									},
								},
							},
						},
						"dnsPolicy":                     "ClusterFirst",
						"restartPolicy":                 "Always",
						"schedulerName":                 "default-scheduler",
						"securityContext":               map[string]interface{}{},
						"terminationGracePeriodSeconds": int64(30),
					},
				},
			},
		},
	}

	result := Normalize(obj)
	spec := result.Object["spec"].(map[string]interface{})

	if spec["replicas"] != int64(2) {
		t.Error("expected replicas to be kept")
	}
	if _, ok := spec["progressDeadlineSeconds"]; ok {
		t.Error("expected progressDeadlineSeconds to be removed")
	}
	if _, ok := spec["revisionHistoryLimit"]; ok {
		t.Error("expected revisionHistoryLimit to be removed")
	}
	if _, ok := spec["strategy"]; ok {
		t.Error("expected default strategy to be removed")
	}

	tmplSpec := spec["template"].(map[string]interface{})["spec"].(map[string]interface{})
	if _, ok := tmplSpec["dnsPolicy"]; ok {
		t.Error("expected dnsPolicy to be removed")
	}
	if _, ok := tmplSpec["restartPolicy"]; ok {
		t.Error("expected restartPolicy to be removed")
	}
	if _, ok := tmplSpec["schedulerName"]; ok {
		t.Error("expected schedulerName to be removed")
	}
	if _, ok := tmplSpec["securityContext"]; ok {
		t.Error("expected empty securityContext to be removed")
	}
	if _, ok := tmplSpec["terminationGracePeriodSeconds"]; ok {
		t.Error("expected terminationGracePeriodSeconds to be removed")
	}

	container := tmplSpec["containers"].([]interface{})[0].(map[string]interface{})
	if _, ok := container["imagePullPolicy"]; ok {
		t.Error("expected default imagePullPolicy to be removed")
	}
	if _, ok := container["terminationMessagePath"]; ok {
		t.Error("expected terminationMessagePath to be removed")
	}
	if _, ok := container["terminationMessagePolicy"]; ok {
		t.Error("expected terminationMessagePolicy to be removed")
	}

	port := container["ports"].([]interface{})[0].(map[string]interface{})
	if _, ok := port["protocol"]; ok {
		t.Error("expected default TCP protocol to be removed")
	}
	if port["containerPort"] != int64(80) {
		t.Error("expected containerPort to be kept")
	}
}

func TestNormalizeDeploymentNonDefaultStrategy(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "app"},
			"spec": map[string]interface{}{
				"strategy": map[string]interface{}{
					"type": "Recreate",
				},
			},
		},
	}

	result := Normalize(obj)
	spec := result.Object["spec"].(map[string]interface{})
	if _, ok := spec["strategy"]; !ok {
		t.Error("expected non-default strategy (Recreate) to be kept")
	}
}

func TestNormalizeServiceDefaults(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata":   map[string]interface{}{"name": "svc"},
			"spec": map[string]interface{}{
				"clusterIP":             "10.96.1.1",
				"clusterIPs":            []interface{}{"10.96.1.1"},
				"internalTrafficPolicy": "Cluster",
				"ipFamilies":            []interface{}{"IPv4"},
				"ipFamilyPolicy":        "SingleStack",
				"sessionAffinity":       "None",
				"type":                  "ClusterIP",
				"selector":              map[string]interface{}{"app": "demo"},
				"ports": []interface{}{
					map[string]interface{}{
						"port":       int64(80),
						"targetPort": int64(80),
						"protocol":   "TCP",
					},
				},
			},
		},
	}

	result := Normalize(obj)
	spec := result.Object["spec"].(map[string]interface{})

	for _, field := range []string{"clusterIP", "clusterIPs", "internalTrafficPolicy", "ipFamilies", "ipFamilyPolicy", "sessionAffinity"} {
		if _, ok := spec[field]; ok {
			t.Errorf("expected %s to be removed", field)
		}
	}
	if spec["type"] != "ClusterIP" {
		t.Error("expected type to be kept")
	}
	if spec["selector"] == nil {
		t.Error("expected selector to be kept")
	}

	port := spec["ports"].([]interface{})[0].(map[string]interface{})
	if _, ok := port["protocol"]; ok {
		t.Error("expected default TCP protocol to be removed from port")
	}
}

func TestNormalizeNamespaceDefaults(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "test-ns",
				"labels": map[string]interface{}{
					"kubernetes.io/metadata.name": "test-ns",
				},
			},
			"spec": map[string]interface{}{
				"finalizers": []interface{}{"kubernetes"},
			},
		},
	}

	result := Normalize(obj)
	metadata := result.Object["metadata"].(map[string]interface{})
	if _, ok := metadata["labels"]; ok {
		t.Error("expected auto-added labels to be removed")
	}
	if _, ok := result.Object["spec"]; ok {
		t.Error("expected spec with only finalizers to be removed")
	}
}

func TestNormalizeImagePullPolicyLatest(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata":   map[string]interface{}{"name": "pod"},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":            "app",
						"image":           "nginx:latest",
						"imagePullPolicy": "Always",
					},
				},
			},
		},
	}

	result := Normalize(obj)
	podSpec := result.Object["spec"].(map[string]interface{})
	container := podSpec["containers"].([]interface{})[0].(map[string]interface{})
	if _, ok := container["imagePullPolicy"]; ok {
		t.Error("expected default Always for :latest to be removed")
	}
}

func TestNormalizeImagePullPolicyNonDefault(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata":   map[string]interface{}{"name": "pod"},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":            "app",
						"image":           "nginx:1.25",
						"imagePullPolicy": "Always",
					},
				},
			},
		},
	}

	result := Normalize(obj)
	podSpec := result.Object["spec"].(map[string]interface{})
	container := podSpec["containers"].([]interface{})[0].(map[string]interface{})
	if container["imagePullPolicy"] != "Always" {
		t.Error("expected non-default Always (for tagged image) to be kept")
	}
}

func TestDefaultImagePullPolicy(t *testing.T) {
	tests := []struct {
		image    string
		expected string
	}{
		{"nginx:1.25", "IfNotPresent"},
		{"nginx:latest", "Always"},
		{"nginx", "Always"},
		{"myregistry/app:v2", "IfNotPresent"},
		{"myregistry/app", "Always"},
		{"", "Always"},
	}
	for _, tt := range tests {
		t.Run(tt.image, func(t *testing.T) {
			if got := defaultImagePullPolicy(tt.image); got != tt.expected {
				t.Errorf("defaultImagePullPolicy(%q) = %q, want %q", tt.image, got, tt.expected)
			}
		})
	}
}

func TestNormalizeJobDefaults(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"metadata":   map[string]interface{}{"name": "my-job"},
			"spec": map[string]interface{}{
				"backoffLimit":   int64(6),
				"completionMode": "NonIndexed",
				"suspend":        false,
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":                     "worker",
								"image":                    "busybox",
								"terminationMessagePath":   "/dev/termination-log",
								"terminationMessagePolicy": "File",
							},
						},
						"restartPolicy": "Never",
					},
				},
			},
		},
	}

	result := Normalize(obj)
	spec := result.Object["spec"].(map[string]interface{})

	if _, ok := spec["backoffLimit"]; ok {
		t.Error("expected backoffLimit to be removed")
	}
	if _, ok := spec["completionMode"]; ok {
		t.Error("expected completionMode to be removed")
	}
	if _, ok := spec["suspend"]; ok {
		t.Error("expected suspend to be removed")
	}
}

func TestNormalizeDaemonSetDefaults(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "DaemonSet",
			"metadata":   map[string]interface{}{"name": "my-ds"},
			"spec": map[string]interface{}{
				"revisionHistoryLimit": int64(10),
				"updateStrategy": map[string]interface{}{
					"type": "RollingUpdate",
				},
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":                     "agent",
								"image":                    "fluentd:v1.16",
								"terminationMessagePath":   "/dev/termination-log",
								"terminationMessagePolicy": "File",
							},
						},
						"dnsPolicy":     "ClusterFirst",
						"restartPolicy": "Always",
					},
				},
			},
		},
	}

	result := Normalize(obj)
	spec := result.Object["spec"].(map[string]interface{})

	if _, ok := spec["revisionHistoryLimit"]; ok {
		t.Error("expected revisionHistoryLimit to be removed")
	}
	if _, ok := spec["updateStrategy"]; ok {
		t.Error("expected default RollingUpdate updateStrategy to be removed")
	}

	tmplSpec := spec["template"].(map[string]interface{})["spec"].(map[string]interface{})
	container := tmplSpec["containers"].([]interface{})[0].(map[string]interface{})
	if _, ok := container["terminationMessagePath"]; ok {
		t.Error("expected terminationMessagePath to be removed")
	}
}

func TestNormalizeDaemonSetNonDefaultStrategy(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "DaemonSet",
			"metadata":   map[string]interface{}{"name": "my-ds"},
			"spec": map[string]interface{}{
				"updateStrategy": map[string]interface{}{
					"type": "OnDelete",
				},
			},
		},
	}

	result := Normalize(obj)
	spec := result.Object["spec"].(map[string]interface{})
	if _, ok := spec["updateStrategy"]; !ok {
		t.Error("expected non-default OnDelete updateStrategy to be kept")
	}
}

func TestNormalizeEmptyResources(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "app"},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":      "app",
								"image":     "nginx:1.25",
								"resources": map[string]interface{}{},
							},
						},
					},
				},
			},
		},
	}

	result := Normalize(obj)
	tmplSpec := result.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
	container := tmplSpec["containers"].([]interface{})[0].(map[string]interface{})
	if _, ok := container["resources"]; ok {
		t.Error("expected empty resources to be removed")
	}
}

func TestRemoveFields(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "test",
				"annotations": map[string]interface{}{
					"keep": "yes",
					"drop": "no",
				},
			},
			"data": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	RemoveFields(obj, []string{"metadata.annotations.drop", "data.key2"})

	metadata := obj.Object["metadata"].(map[string]interface{})
	annotations := metadata["annotations"].(map[string]interface{})
	if _, ok := annotations["drop"]; ok {
		t.Error("expected 'drop' annotation to be removed")
	}
	if annotations["keep"] != "yes" {
		t.Error("expected 'keep' annotation to remain")
	}

	data := obj.Object["data"].(map[string]interface{})
	if _, ok := data["key2"]; ok {
		t.Error("expected key2 to be removed")
	}
	if data["key1"] != "value1" {
		t.Error("expected key1 to remain")
	}
}

func TestRemoveFieldsTopLevel(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata":   map[string]interface{}{"name": "test"},
			"data":       map[string]interface{}{"key": "val"},
		},
	}

	RemoveFields(obj, []string{"data"})

	if _, ok := obj.Object["data"]; ok {
		t.Error("expected 'data' to be removed")
	}
}

func TestRemoveFieldsNonExistent(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata":   map[string]interface{}{"name": "test"},
		},
	}

	// Should not panic
	RemoveFields(obj, []string{"nonexistent.field.path"})
}

func TestRemoveFieldsNil(t *testing.T) {
	// Should not panic
	RemoveFields(nil, []string{"some.field"})
}

func TestRemoveFieldsCleansEmptyParent(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata":   map[string]interface{}{"name": "test"},
			"data": map[string]interface{}{
				"only-key": "val",
			},
		},
	}

	RemoveFields(obj, []string{"data.only-key"})

	// data should be cleaned up since it's now empty
	if _, ok := obj.Object["data"]; ok {
		t.Error("expected empty 'data' parent to be removed")
	}
}

func TestNormalizeStatefulSetUpdateStrategy(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "StatefulSet",
			"metadata":   map[string]interface{}{"name": "my-sts"},
			"spec": map[string]interface{}{
				"revisionHistoryLimit": int64(10),
				"updateStrategy": map[string]interface{}{
					"type": "RollingUpdate",
				},
			},
		},
	}

	result := Normalize(obj)
	spec := result.Object["spec"].(map[string]interface{})

	if _, ok := spec["updateStrategy"]; ok {
		t.Error("expected default RollingUpdate updateStrategy to be removed from StatefulSet")
	}
}

func TestNormalizeStatefulSetNonDefaultUpdateStrategy(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "StatefulSet",
			"metadata":   map[string]interface{}{"name": "my-sts"},
			"spec": map[string]interface{}{
				"updateStrategy": map[string]interface{}{
					"type": "OnDelete",
				},
			},
		},
	}

	result := Normalize(obj)
	spec := result.Object["spec"].(map[string]interface{})

	if _, ok := spec["updateStrategy"]; !ok {
		t.Error("expected non-default OnDelete updateStrategy to be kept")
	}
}

func TestNormalizeInitContainers(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "app"},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "app",
								"image": "nginx:1.25",
							},
						},
						"initContainers": []interface{}{
							map[string]interface{}{
								"name":                     "init",
								"image":                    "busybox:1.36",
								"terminationMessagePath":   "/dev/termination-log",
								"terminationMessagePolicy": "File",
								"imagePullPolicy":          "IfNotPresent",
								"resources":                map[string]interface{}{},
							},
						},
					},
				},
			},
		},
	}

	result := Normalize(obj)
	tmplSpec := result.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
	initContainer := tmplSpec["initContainers"].([]interface{})[0].(map[string]interface{})

	if _, ok := initContainer["terminationMessagePath"]; ok {
		t.Error("expected terminationMessagePath to be removed from initContainer")
	}
	if _, ok := initContainer["terminationMessagePolicy"]; ok {
		t.Error("expected terminationMessagePolicy to be removed from initContainer")
	}
	if _, ok := initContainer["imagePullPolicy"]; ok {
		t.Error("expected default imagePullPolicy to be removed from initContainer")
	}
	if _, ok := initContainer["resources"]; ok {
		t.Error("expected empty resources to be removed from initContainer")
	}
}

func TestNormalizePodTemplateNoSpec(t *testing.T) {
	// template without spec should not panic
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "app"},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{"app": "test"},
					},
				},
			},
		},
	}

	// Should not panic
	result := Normalize(obj)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestNormalizePodTemplateCreationTimestamp(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "app"},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": nil,
						"labels":            map[string]interface{}{"app": "test"},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "app",
								"image": "nginx:1.25",
							},
						},
					},
				},
			},
		},
	}

	result := Normalize(obj)
	tmplMeta := result.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})
	if _, ok := tmplMeta["creationTimestamp"]; ok {
		t.Error("expected template creationTimestamp to be removed")
	}
	if tmplMeta["labels"] == nil {
		t.Error("expected labels to be preserved")
	}
}

func TestNormalizeNonSecurityContext(t *testing.T) {
	// Non-empty securityContext should be kept
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata":   map[string]interface{}{"name": "pod"},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "app",
						"image": "nginx:1.25",
					},
				},
				"securityContext": map[string]interface{}{
					"runAsUser": int64(1000),
				},
			},
		},
	}

	result := Normalize(obj)
	podSpec := result.Object["spec"].(map[string]interface{})
	if _, ok := podSpec["securityContext"]; !ok {
		t.Error("expected non-empty securityContext to be kept")
	}
}

func TestNormalizeDoesNotModifyOriginal(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":            "test",
				"uid":             "abc-123",
				"resourceVersion": "999",
			},
			"status": map[string]interface{}{
				"phase": "Active",
			},
		},
	}

	_ = Normalize(obj)

	// Original should be untouched
	if _, ok := obj.Object["status"]; !ok {
		t.Error("original object status should not be modified")
	}
	metadata := obj.Object["metadata"].(map[string]interface{})
	if _, ok := metadata["uid"]; !ok {
		t.Error("original object uid should not be modified")
	}
}
