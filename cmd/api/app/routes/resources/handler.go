package resources

import (
	"github.com/gin-gonic/gin"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	query "github.com/karmada-io/dashboard/pkg/resource/common"
	"github.com/karmada-io/dashboard/pkg/resource/cronjob"
	"github.com/karmada-io/dashboard/pkg/resource/daemonset"
	"github.com/karmada-io/dashboard/pkg/resource/deployment"
	"github.com/karmada-io/dashboard/pkg/resource/job"
	"github.com/karmada-io/dashboard/pkg/resource/namespace"
	"github.com/karmada-io/dashboard/pkg/resource/statefulset"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"strings"
)

const (
	ResourceNamesKey      = "names"
	ResourceLabelsKey     = "labels"
	ResourceNamespacesKey = "namespaces"
)

func HandleGetResourceNames(c *gin.Context) {
	ns := common.ParseNamespaceQuery(c)
	kind := c.Query("kind")
	k8sClient := client.InClusterClientForKarmadaAPIServer()

	resourceHandlers := map[string]func(kubernetes.Interface, *query.NamespaceQuery) ([]string, error){
		/*		"namespace": func(c kubernetes.Interface, _ *query.NamespaceQuery) ([]string, error) {
				return namespace.GetNamespaceNames(c)
			},*/
		"deployment":  deployment.GetDeploymentNames,
		"statefulset": statefulset.GetStatefulSetNames,
		"daemonset":   daemonset.GetDaemonSetNames,
		"cronjob":     cronjob.GetCronJobNames,
		"job":         job.GetJobNames,
	}

	handler, exists := resourceHandlers[kind]
	if !exists {
		response.FailedWithError(c, apperrors.UnsupportedResourceKind)
		return
	}

	result, err := handler(k8sClient, ns)
	if err != nil {
		response.FailedWithError(c, err)
		return
	}

	if result == nil {
		result = []string{}
	}
	response.Success(c, gin.H{ResourceNamesKey: result})
}

func HandleGetResourceLabels(c *gin.Context) {
	ns := common.ParseNamespaceQuery(c)
	kind := c.Query("kind")
	k8sClient := client.InClusterClientForKarmadaAPIServer()

	resourceHandlers := map[string]func(kubernetes.Interface, *query.NamespaceQuery) ([]string, error){
		/*		"namespace": func(c kubernetes.Interface, _ *query.NamespaceQuery) ([]string, error) {
				return namespace.GetNamespaceLabelSelectors(c)
			},*/
		"deployment":  deployment.GetDeploymentLabelSelectors,
		"statefulset": statefulset.GetStatefulSetLabelSelectors,
		"daemonset":   daemonset.GetDaemonSetLabelSelectors,
		"cronjob":     cronjob.GetCronJobLabelSelectors,
		"job":         job.GetJobLabelSelectors,
	}

	handler, exists := resourceHandlers[kind]
	if !exists {
		response.FailedWithError(c, apperrors.UnsupportedResourceKind)
		return
	}

	result, err := handler(k8sClient, ns)
	if err != nil {
		response.FailedWithError(c, err)
		return
	}

	if result == nil {
		result = []string{}
	}

	response.Success(c, gin.H{ResourceLabelsKey: result})
}

func HandleGetResourceNamespace(c *gin.Context) {
	kind := c.Query("kind")
	ns := query.NewNamespaceQuery([]string{})
	k8sClient := client.InClusterClientForKarmadaAPIServer()

	resourceHandlers := map[string]func(kubernetes.Interface, *query.NamespaceQuery) ([]string, error){
		"deployment":  deployment.GetNamespacesWithDeployment,
		"statefulset": statefulset.GetNamespacesWithStatefulSet,
		"daemonset":   daemonset.GetNamespacesWithDaemonSet,
		"cronjob":     cronjob.GetNamespacesWithCronJob,
		"job":         job.GetNamespacesWithJob,
		"": func(c kubernetes.Interface, _ *query.NamespaceQuery) ([]string, error) {
			return namespace.GetNamespaceNames(c)
		},
	}

	handler, exists := resourceHandlers[kind]
	if !exists {
		response.FailedWithError(c, apperrors.UnsupportedResourceKind)
		return
	}

	result, err := handler(k8sClient, ns)
	if err != nil {
		response.FailedWithError(c, err)
		return
	}

	//result = internalCommon.FilterNamespaceByArrayString(result)

	if result == nil {
		result = []string{}
	}
	response.Success(c, gin.H{ResourceNamespacesKey: result})
}

// ------------------------ Create / Update ------------------------

func HandleCreateResource(c *gin.Context) {
	handleVerberActionWithManifest(c, false, func(verber client.ResourceVerber, obj *unstructured.Unstructured) error {
		_, err := verber.Create(obj)
		return err
	}, response.Created)
}

func HandleUpdateResource(c *gin.Context) {
	handleVerberActionWithManifest(c, true, func(verber client.ResourceVerber, obj *unstructured.Unstructured) error {
		return verber.Update(obj)
	}, func(c *gin.Context) {
		response.SuccessWithMessage(c, msgkey.ResourceUpdateSuccess)
	})
}

// ------------------------ Get / Delete ------------------------

func HandleGetResourceYaml(c *gin.Context) {
	handleVerberActionWithPathParam(c, func(verber client.ResourceVerber, kind, ns, name string) (interface{}, error) {
		return verber.Get(kind, ns, name)
	}, func(c *gin.Context, obj interface{}) {
		u, ok := obj.(*unstructured.Unstructured)
		if !ok {
			response.FailedWithError(c, apperrors.FailedRequest)
			return
		}

		yamlStr, err := common.EncodeToYAML(u, k8sscheme.Scheme)
		if err != nil {
			response.FailedWithError(c, err)
			return
		}

		result := &v1.ResourceYaml{
			Namespace: u.GetNamespace(),
			Name:      u.GetName(),
			UID:       u.GetUID(),
			Yaml:      yamlStr,
		}
		response.Success(c, result)
	})
}

func HandleDeleteResource(c *gin.Context) {
	handleVerberActionWithPathParam(c, func(verber client.ResourceVerber, kind, ns, name string) (interface{}, error) {
		err := verber.Delete(kind, ns, name, true)
		return nil, err
	}, func(c *gin.Context, _ interface{}) {
		response.SuccessWithMessage(c, msgkey.ResourceDeletionCompleted)
	})
}

func HandleListResource(c *gin.Context) {
	dataSelect := common.ParseDataSelectPathParameter(c)
	if err := common.IsValidProperties(dataSelect); err != nil {
		response.FailedWithError(c, err)
		return
	}

	kind := c.Param("kind")
	if !IsKindSupportedForPathParam(kind) {
		response.FailedWithError(c, apperrors.UnsupportedResourceKind)
		return
	}

	verber, err := client.VerberClient(c.Request)
	if err != nil {
		klog.ErrorS(err, "Failed to init verber client")
		response.FailedWithError(c, err)
		return
	}

	namespace := common.ParseNamespaceQuery(c)
	result, err := GetResourceList(verber, strings.ToLower(kind), namespace, dataSelect)
	if err != nil {
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, result)
}
