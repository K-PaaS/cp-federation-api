package common

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/pkg/dataselect"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog/v2"
	"math"
	"sort"
	"strings"
)

func GetClaims(c *gin.Context) jwt.MapClaims {
	val, exists := c.Get("claims")
	if !exists {
		return jwt.MapClaims{}
	}
	claims, exists := val.(jwt.MapClaims)
	if !exists {
		return jwt.MapClaims{}
	}
	return claims
}

func IsValidProperties(dataSelect *dataselect.DataSelectQuery) error {
	// check sort option
	sortQuery := dataSelect.SortQuery
	if sortQuery.HasInvalidAscending {
		return apperrors.InvalidSortOrder
	}
	if sortQuery.HasInvalidProperty {
		return apperrors.InvalidSortProperty
	}
	return nil
}

func RoundToTwoDecimals(val float64) float64 {
	//fmt.Println("val*100:", val*100)
	//fmt.Println("math.Round(val*100):", math.Round(val*100))
	//fmt.Println("math.Round(val*100) / 100:", math.Round(val*100)/100)
	return math.Round(val*100) / 100
}

func EncodeToYAMLString(obj runtime.Object, scheme *runtime.Scheme) (string, error) {
	var buf bytes.Buffer

	serializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme,
		scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: false,
			Strict: false,
		},
	)

	if err := serializer.Encode(obj, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func SetGVKFromScheme(obj runtime.Object, scheme *runtime.Scheme) error {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	if len(gvks) > 0 {
		fmt.Println("gvks", gvks)
		obj.GetObjectKind().SetGroupVersionKind(gvks[0])
	}
	return nil
}

func EncodeToYAML(obj runtime.Object, scheme *runtime.Scheme) (string, error) {
	err := SetGVKFromScheme(obj, scheme)
	if err != nil {
		klog.ErrorS(err, "failed to set gvk from scheme")
		return "", err
	}

	yaml, err := EncodeToYAMLString(obj, scheme)
	if err != nil {
		klog.ErrorS(err, "failed to encode to yaml string")
		return "", err
	}

	return yaml, nil
}

func ExtractLabels[T metav1.Object](items []T) []string {
	labelMap := make(map[string]map[string]struct{})

	for _, item := range items {
		labels := item.GetLabels()
		for k, v := range labels {
			if _, exists := labelMap[k]; !exists {
				labelMap[k] = make(map[string]struct{})
			}
			labelMap[k][v] = struct{}{}
		}
	}

	var labelSelectors []string
	for k, values := range labelMap {
		for v := range values {
			labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
		}
	}

	sort.Strings(labelSelectors)
	return labelSelectors
}

func ExtractNamespaces[T metav1.Object](items []T) []string {
	nsSet := make(map[string]struct{})
	for _, item := range items {
		if ns := item.GetNamespace(); ns != "" {
			nsSet[ns] = struct{}{}
		}
	}
	var namespaces []string
	for ns := range nsSet {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)
	return namespaces
}

// ParseKeyValueStrings converts an array of "key＝value" strings into a map[string]string.
func ParseKeyValueStrings(labelStrings []string) (map[string]string, error) {
	labels := make(map[string]string)
	for _, item := range labelStrings {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, apperrors.InvalidKeyValueFormat
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, apperrors.InvalidKeyValueFormat
		}
		labels[key] = value
	}
	return labels, nil
}
