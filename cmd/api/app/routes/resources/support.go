package resources

import (
	"github.com/gin-gonic/gin"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
	"strings"
)

var supportedKinds = map[string]string{
	"deployment":  "Deployment",
	"statefulset": "StatefulSet",
	"daemonset":   "DaemonSet",
	"cronjob":     "CronJob",
	"job":         "Job",
	"configmap":   "ConfigMap",
	"secret":      "Secret",
}

const defaultNamespace = "default"

// ------------------------ Create / Update ------------------------
func handleVerberActionWithManifest(
	c *gin.Context,
	isUpdate bool,
	action func(verber client.ResourceVerber, obj *unstructured.Unstructured) error,
	onSuccess func(*gin.Context),
) {
	verber, err := client.VerberClient(c.Request)
	if err != nil {
		klog.ErrorS(err, "Failed to init verber client")
		response.FailedWithError(c, err)
		return
	}

	obj, err := parseManifestStrict(c)
	if err != nil {
		klog.ErrorS(err, "Failed to parse manifest")
		response.FailedWithError(c, err)
		return
	}

	if err := validatePreconditions(verber, obj, isUpdate); err != nil {
		response.FailedWithError(c, err)
		return
	}

	if err := action(verber, obj); err != nil {
		klog.ErrorS(err, "Verber action failed")
		response.FailedWithError(c, apperrors.ResourceError(err))
		return
	}

	onSuccess(c)
}

func parseManifestStrict(c *gin.Context) (*unstructured.Unstructured, error) {
	var manifest v1.ManifestRequest
	if err := c.ShouldBindJSON(&manifest); err != nil {
		return nil, apperrors.RequestValueInvalid
	}

	jsonBytes, err := yaml.YAMLToJSON([]byte(manifest.Data))
	if err != nil {
		return nil, apperrors.InvalidYamlFormat
	}

	var obj unstructured.Unstructured
	if err := obj.UnmarshalJSON(jsonBytes); err != nil {
		return nil, apperrors.InvalidYamlFormat
	}

	kind := obj.GetKind()
	if !IsKindStrictlySupported(kind) {
		return nil, apperrors.UnsupportedResourceKind
	}

	if obj.GetNamespace() == "" {
		klog.Infof("No namespace specified for kind=%s, name=%s. Defaulting to '%s' namespace.",
			obj.GetKind(), obj.GetName(), defaultNamespace)
		obj.SetNamespace(defaultNamespace)
	}

	return &obj, nil
}

func IsKindStrictlySupported(kind string) bool {
	expected, ok := supportedKinds[strings.ToLower(kind)]
	return ok && kind == expected
}

func validatePreconditions(verber client.ResourceVerber, obj *unstructured.Unstructured, isUpdate bool) error {
	if isUpdate {
		//  check request resource existence for update
		_, err := verber.Get(strings.ToLower(obj.GetKind()), obj.GetNamespace(), obj.GetName())
		if err != nil {
			klog.ErrorS(err, "Failed to check resource existence")
			return apperrors.ResourceError(err)
		}
	} else {
		// check request namespace existence for create
		_, err := verber.Get("namespace", "", obj.GetNamespace())
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.InfoS("Namespace not found for create", "namespace", obj.GetNamespace())
				return apperrors.NamespaceNotFound
			}
			klog.ErrorS(err, "Failed to check namespace existence")
			return apperrors.ResourceError(err)
		}
	}
	return nil
}

// ------------------------ Get / Delete ------------------------
func handleVerberActionWithPathParam(
	c *gin.Context,
	action func(verber client.ResourceVerber, kind, namespace, name string) (interface{}, error),
	onSuccess func(*gin.Context, interface{}),
) {
	verber, err := client.VerberClient(c.Request)
	if err != nil {
		klog.ErrorS(err, "Failed to init verber client")
		response.FailedWithError(c, err)
		return
	}

	kind, namespace, name := c.Param("kind"), c.Param("namespace"), c.Param("name")
	if err := validateRequestParams(kind, namespace, name); err != nil {
		response.FailedWithError(c, err)
		return
	}

	obj, err := action(verber, strings.ToLower(kind), namespace, name)
	if err != nil {
		klog.ErrorS(err, "Verber action failed", "kind", kind, "namespace", namespace, "name", name)
		response.FailedWithError(c, apperrors.ResourceError(err))
		return
	}

	onSuccess(c, obj)
}

func validateRequestParams(kind, namespace, name string) error {
	if kind == "" || namespace == "" || name == "" {
		return apperrors.RequestValueInvalid
	}

	if !IsKindSupportedForPathParam(kind) {
		return apperrors.UnsupportedResourceKind
	}

	return nil
}

func IsKindSupportedForPathParam(kind string) bool {
	_, ok := supportedKinds[strings.ToLower(kind)]
	return ok
}
