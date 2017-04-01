package kubernetes

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"encoding/base64"
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/kubernetes/pkg/api/resource"
	api "k8s.io/kubernetes/pkg/api/v1"
)

func idParts(id string) (string, string) {
	parts := strings.Split(id, "/")
	return parts[0], parts[1]
}

func buildId(meta api.ObjectMeta) string {
	return meta.Namespace + "/" + meta.Name
}

func expandMetadata(in []interface{}) api.ObjectMeta {
	meta := api.ObjectMeta{}
	if len(in) < 1 {
		return meta
	}
	m := in[0].(map[string]interface{})

	meta.Annotations = filterAnnotations(expandStringMap(m["annotations"].(map[string]interface{})))
	meta.Labels = expandStringMap(m["labels"].(map[string]interface{}))

	if v, ok := m["generate_name"]; ok {
		meta.GenerateName = v.(string)
	}
	if v, ok := m["name"]; ok {
		meta.Name = v.(string)
	}
	if v, ok := m["namespace"]; ok {
		meta.Namespace = v.(string)
	}

	return meta
}

func patchMetadata(keyPrefix, pathPrefix string, d *schema.ResourceData) PatchOperations {
	ops := make([]PatchOperation, 0, 0)
	if d.HasChange(keyPrefix + "annotations") {
		oldV, newV := d.GetChange(keyPrefix + "annotations")
		diffOps := diffStringMap(pathPrefix+"annotations", oldV.(map[string]interface{}), newV.(map[string]interface{}))
		ops = append(ops, diffOps...)
	}
	if d.HasChange(keyPrefix + "labels") {
		oldV, newV := d.GetChange(keyPrefix + "labels")
		diffOps := diffStringMap(pathPrefix+"labels", oldV.(map[string]interface{}), newV.(map[string]interface{}))
		ops = append(ops, diffOps...)
	}
	return ops
}

func expandStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v.(string)
	}
	return result
}

func flattenMetadata(meta api.ObjectMeta) []map[string]interface{} {
	m := make(map[string]interface{})
	m["annotations"] = filterAnnotations(meta.Annotations)
	if meta.GenerateName != "" {
		m["generate_name"] = meta.GenerateName
	}
	m["labels"] = meta.Labels
	m["name"] = meta.Name
	m["resource_version"] = meta.ResourceVersion
	m["self_link"] = meta.SelfLink
	m["uid"] = fmt.Sprintf("%v", meta.UID)
	m["generation"] = meta.Generation

	if meta.Namespace != "" {
		m["namespace"] = meta.Namespace
	}

	return []map[string]interface{}{m}
}

func filterAnnotations(m map[string]string) map[string]string {
	for k, _ := range m {
		if isInternalAnnotationKey(k) {
			delete(m, k)
		}
	}
	return m
}

func byteMapToStringMap(m map[string][]byte) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = string(v)
	}
	return result
}

func isInternalAnnotationKey(annotationKey string) bool {
	u, err := url.Parse("//" + annotationKey)
	if err == nil && strings.HasSuffix(u.Hostname(), "kubernetes.io") {
		return true
	}

	return false
}

func ptrToString(s string) *string {
	return &s
}

func ptrToInt(i int) *int {
	return &i
}

func ptrToBool(b bool) *bool {
	return &b
}

func ptrToInt32(i int32) *int32 {
	return &i
}

func sliceOfString(slice []interface{}) []string {
	result := make([]string, len(slice), len(slice))
	for i, s := range slice {
		result[i] = s.(string)
	}
	return result
}

func base64EncodeStringMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		value := v.(string)
		result[k] = (base64.StdEncoding.EncodeToString([]byte(value)))
	}
	return result
}

func flattenResourceList(l api.ResourceList) map[string]string {
	log.Printf("[DEBUG] Flattening resource list - IN: %#v", l)
	m := make(map[string]string)
	for k, v := range l {
		m[string(k)] = v.String()
	}
	log.Printf("[DEBUG] Flattening resource list - OUT: %#v", m)
	return m
}

func expandMapToResourceList(m map[string]interface{}) (api.ResourceList, error) {
	log.Printf("[DEBUG] Expanding resource list - IN: %#v", m)
	out := make(map[api.ResourceName]resource.Quantity)
	for stringKey, v := range m {
		key := api.ResourceName(stringKey)
		value, err := resource.ParseQuantity(v.(string))
		if err != nil {
			return out, err
		}

		out[key] = value
	}
	log.Printf("[DEBUG] Expanding resource list - OUT: %#v", out)
	return out, nil
}

func ifaceFromString(s string) interface{} {
	return s
}

func flattenPersistentVolumeAccessModes(in []api.PersistentVolumeAccessMode) *schema.Set {
	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	return schema.NewSet(schema.HashString, out)
}

func expandPersistentVolumeAccessModes(s []interface{}) []api.PersistentVolumeAccessMode {
	out := make([]api.PersistentVolumeAccessMode, len(s), len(s))
	for i, v := range s {
		out[i] = api.PersistentVolumeAccessMode(v.(string))
	}
	return out
}

func ptrToAzureDataDiskCachingMode(in api.AzureDataDiskCachingMode) *api.AzureDataDiskCachingMode {
	return &in
}
