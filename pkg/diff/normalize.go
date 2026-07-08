package diff

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Normalize removes cluster-managed fields from a resource for clean comparison.
func Normalize(obj *unstructured.Unstructured) *unstructured.Unstructured {
	if obj == nil {
		return nil
	}

	normalized := obj.DeepCopy()

	// Remove top-level status
	delete(normalized.Object, "status")

	// Remove metadata fields
	if metadata, ok := normalized.Object["metadata"].(map[string]interface{}); ok {
		delete(metadata, "managedFields")
		delete(metadata, "resourceVersion")
		delete(metadata, "uid")
		delete(metadata, "creationTimestamp")
		delete(metadata, "generation")
		delete(metadata, "selfLink")

		// Remove specific annotations
		if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
			delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
			delete(annotations, "deployment.kubernetes.io/revision")
			if len(annotations) == 0 {
				delete(metadata, "annotations")
			}
		}

		// Remove auto-added labels
		if labels, ok := metadata["labels"].(map[string]interface{}); ok {
			delete(labels, "kubernetes.io/metadata.name")
			if len(labels) == 0 {
				delete(metadata, "labels")
			}
		}
	}

	// Remove kind-specific cluster defaults
	kind := normalized.GetKind()
	spec, _ := normalized.Object["spec"].(map[string]interface{})

	switch kind {
	case "Deployment", "StatefulSet":
		normalizeDeploymentSpec(spec)
	case "Service":
		normalizeServiceSpec(spec)
	case "Namespace":
		normalizeNamespaceSpec(normalized.Object)
	case "Pod":
		normalizePodSpec(spec)
	case "Job":
		normalizeJobSpec(spec)
	case "DaemonSet":
		normalizeDaemonSetSpec(spec)
	}

	return normalized
}

// RemoveFields removes the specified field paths from an unstructured object.
// Field paths use dot notation, e.g., "metadata.annotations.some-key", "spec.replicas".
func RemoveFields(obj *unstructured.Unstructured, fields []string) {
	if obj == nil {
		return
	}
	for _, field := range fields {
		removeNestedField(obj.Object, strings.Split(field, "."))
	}
}

// removeNestedField removes a field at the given path from a nested map.
func removeNestedField(obj map[string]interface{}, path []string) {
	if len(path) == 0 || obj == nil {
		return
	}
	if len(path) == 1 {
		delete(obj, path[0])
		return
	}
	next, ok := obj[path[0]].(map[string]interface{})
	if !ok {
		return
	}
	removeNestedField(next, path[1:])
	// Clean up empty parent maps
	if len(next) == 0 {
		delete(obj, path[0])
	}
}

// normalizeDeploymentSpec removes Kubernetes-defaulted fields from Deployment/StatefulSet spec.
func normalizeDeploymentSpec(spec map[string]interface{}) {
	if spec == nil {
		return
	}
	delete(spec, "progressDeadlineSeconds")
	delete(spec, "revisionHistoryLimit")

	// Remove default strategy
	if strategy, ok := spec["strategy"].(map[string]interface{}); ok {
		if strategy["type"] == "RollingUpdate" {
			delete(spec, "strategy")
		}
	}

	// Remove updateStrategy default for StatefulSet
	if us, ok := spec["updateStrategy"].(map[string]interface{}); ok {
		if us["type"] == "RollingUpdate" {
			delete(spec, "updateStrategy")
		}
	}

	normalizePodTemplate(spec)
}

// normalizeDaemonSetSpec removes Kubernetes-defaulted fields from DaemonSet spec.
func normalizeDaemonSetSpec(spec map[string]interface{}) {
	if spec == nil {
		return
	}
	delete(spec, "revisionHistoryLimit")

	if us, ok := spec["updateStrategy"].(map[string]interface{}); ok {
		if us["type"] == "RollingUpdate" {
			delete(spec, "updateStrategy")
		}
	}

	normalizePodTemplate(spec)
}

// normalizeJobSpec removes Kubernetes-defaulted fields from Job spec.
func normalizeJobSpec(spec map[string]interface{}) {
	if spec == nil {
		return
	}
	delete(spec, "backoffLimit")
	delete(spec, "completionMode")
	delete(spec, "suspend")

	normalizePodTemplate(spec)
}

// normalizePodTemplate handles the template.spec.containers defaults.
func normalizePodTemplate(spec map[string]interface{}) {
	tmpl, ok := spec["template"].(map[string]interface{})
	if !ok {
		return
	}

	// Remove template.metadata.creationTimestamp (Kubernetes sets this to null)
	if metadata, ok := tmpl["metadata"].(map[string]interface{}); ok {
		delete(metadata, "creationTimestamp")
	}

	podSpec, ok := tmpl["spec"].(map[string]interface{})
	if !ok {
		return
	}
	normalizePodSpec(podSpec)
}

// normalizePodSpec removes Kubernetes-defaulted fields from a pod spec.
func normalizePodSpec(podSpec map[string]interface{}) {
	if podSpec == nil {
		return
	}
	delete(podSpec, "dnsPolicy")
	delete(podSpec, "restartPolicy")
	delete(podSpec, "schedulerName")
	delete(podSpec, "terminationGracePeriodSeconds")

	// Remove empty securityContext
	if sc, ok := podSpec["securityContext"].(map[string]interface{}); ok && len(sc) == 0 {
		delete(podSpec, "securityContext")
	}

	// Normalize containers
	normalizeContainers(podSpec, "containers")
	normalizeContainers(podSpec, "initContainers")
}

// normalizeContainers removes defaulted fields from container specs.
func normalizeContainers(podSpec map[string]interface{}, key string) {
	containers, ok := podSpec[key].([]interface{})
	if !ok {
		return
	}
	for _, c := range containers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		delete(container, "terminationMessagePath")
		delete(container, "terminationMessagePolicy")

		// Remove empty resources (cluster adds resources: {} by default)
		if res, ok := container["resources"].(map[string]interface{}); ok && len(res) == 0 {
			delete(container, "resources")
		}

		// Remove default imagePullPolicy
		if container["imagePullPolicy"] == "IfNotPresent" || container["imagePullPolicy"] == "Always" {
			// Only remove if it matches the Kubernetes default for the tag
			image, _ := container["image"].(string)
			defaultPolicy := defaultImagePullPolicy(image)
			if container["imagePullPolicy"] == defaultPolicy {
				delete(container, "imagePullPolicy")
			}
		}

		// Remove default protocol from ports
		if ports, ok := container["ports"].([]interface{}); ok {
			for _, p := range ports {
				port, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				if port["protocol"] == "TCP" {
					delete(port, "protocol")
				}
			}
		}
	}
}

// defaultImagePullPolicy returns what Kubernetes defaults to for a given image.
func defaultImagePullPolicy(image string) string {
	// Kubernetes defaults to Always for :latest or no tag, IfNotPresent otherwise
	if image == "" {
		return "Always"
	}
	// Check for :latest or no tag
	for i := len(image) - 1; i >= 0; i-- {
		if image[i] == ':' {
			if image[i+1:] == "latest" {
				return "Always"
			}
			return "IfNotPresent"
		}
		if image[i] == '/' {
			break
		}
	}
	// No tag → defaults to Always
	return "Always"
}

// normalizeServiceSpec removes Kubernetes-defaulted fields from Service spec.
func normalizeServiceSpec(spec map[string]interface{}) {
	if spec == nil {
		return
	}
	delete(spec, "clusterIP")
	delete(spec, "clusterIPs")
	delete(spec, "internalTrafficPolicy")
	delete(spec, "ipFamilies")
	delete(spec, "ipFamilyPolicy")
	delete(spec, "sessionAffinity")

	// Remove default protocol from ports
	if ports, ok := spec["ports"].([]interface{}); ok {
		for _, p := range ports {
			port, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			if port["protocol"] == "TCP" {
				delete(port, "protocol")
			}
		}
	}
}

// normalizeNamespaceSpec removes Kubernetes-defaulted fields from Namespace.
func normalizeNamespaceSpec(obj map[string]interface{}) {
	// Remove spec.finalizers (auto-added by Kubernetes)
	if spec, ok := obj["spec"].(map[string]interface{}); ok {
		delete(spec, "finalizers")
		if len(spec) == 0 {
			delete(obj, "spec")
		}
	}
}
