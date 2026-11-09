package diff

import (
	"encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Placeholders for Secret values. No hash is shown: a short password falls to a
// dictionary attack on its digest.
const (
	Redacted        = "<redacted>"
	RedactedChanged = "<redacted, changed>"
)

func isSecret(obj *unstructured.Unstructured) bool {
	return obj != nil && obj.GetKind() == "Secret" && obj.GetAPIVersion() == "v1"
}

// foldStringData merges stringData into data the way the API server does on
// write, so a manifest that uses stringData compares equal to the stored Secret.
func foldStringData(obj *unstructured.Unstructured) {
	stringData, ok := obj.Object["stringData"].(map[string]interface{})
	if !ok {
		return
	}
	data, _ := obj.Object["data"].(map[string]interface{})
	if data == nil {
		data = make(map[string]interface{}, len(stringData))
	}
	for key, value := range stringData {
		data[key] = base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(value)))
	}
	obj.Object["data"] = data
	delete(obj.Object, "stringData")
}

// redactSecretData replaces every data value on both sides, marking the local
// keys whose value differs from the cluster's so the diff still names them.
func redactSecretData(local, cluster *unstructured.Unstructured) {
	localData, _ := local.Object["data"].(map[string]interface{})
	clusterData, _ := cluster.Object["data"].(map[string]interface{})
	for key, value := range localData {
		if clusterValue, ok := clusterData[key]; ok && clusterValue != value {
			localData[key] = RedactedChanged
		} else {
			localData[key] = Redacted
		}
	}
	for key := range clusterData {
		clusterData[key] = Redacted
	}
}
