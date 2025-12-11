package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/adapter/federation"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
)

type Handler struct {
	Adapter federation.FederationAdapter
}

func NewHandler(adapter federation.FederationAdapter) *Handler {
	return &Handler{
		Adapter: adapter,
	}
}

type SyncResource struct {
	Name         string `json:"name"`
	IsDuplicated bool   `json:"isDuplicated"`
}

type SyncRequest struct {
	CreateNamespaces []string          `json:"createNamespace"`
	Data             []SyncRequestData `json:"data"`
}

type SyncResponse struct {
	TotalResource   int `json:"totalResource"`
	FailResource    int `json:"failResource"`
	SuccessResource int `json:"successResource"`
}

type SyncRequestData struct {
	Namespace string                `json:"namespace"`
	List      []SyncRequestResource `json:"list"`
}
type SyncRequestResource struct {
	Kind string   `json:"kind"`
	List []string `json:"list"`
}

func BuildSyncResourceList[T NamedResource](clusterItems, karmadaItems []T) []SyncResource {
	var result []SyncResource
	for _, c := range clusterItems {
		isDuplicated := false
		for _, k := range karmadaItems {
			if c.GetName() == k.GetName() {
				isDuplicated = true
				break
			}
		}
		result = append(result, SyncResource{
			Name:         c.GetName(),
			IsDuplicated: isDuplicated,
		})
	}
	return result
}

func ToPtrSlice[T any](items []T) []*T {
	ptrs := make([]*T, 0, len(items))
	for i := range items {
		ptrs = append(ptrs, &items[i])
	}
	return ptrs
}

var kindToGVR = map[string]schema.GroupVersionResource{
	"namespace": {
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	},
	"deployment": {
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	},
	"statefulset": {
		Group:    "apps",
		Version:  "v1",
		Resource: "statefulsets",
	},
	"daemonset": {
		Group:    "apps",
		Version:  "v1",
		Resource: "daemonsets",
	},
	"job": {
		Group:    "batch",
		Version:  "v1",
		Resource: "jobs",
	},
	"cronjob": {
		Group:    "batch",
		Version:  "v1",
		Resource: "cronjobs",
	},
	"configmap": {
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	},
	"secret": {
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	},
}

func listK8sResources(kind string, namespace string, dynClient dynamic.Interface) ([]unstructured.Unstructured, error) {
	gvr, ok := kindToGVR[kind]
	if !ok {
		return nil, nil
	}

	var resources *unstructured.UnstructuredList
	var err error
	if kind == "namespace" {
		resources, err = dynClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	} else {
		if namespace == "" {
			return nil, fmt.Errorf("require namespace")
		}
		resources, err = dynClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}

	return resources.Items, nil
}

type NamedResource interface {
	GetName() string
}

func ConvertUnstructuredList[T any](items []unstructured.Unstructured) ([]T, error) {
	var results []T
	for _, item := range items {
		var result T
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert: %w", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (h Handler) HandleGetSyncResources(c *gin.Context) {
	kind := c.Query("kind")

	if _, ok := kindToGVR[kind]; !ok {
		log.Printf("not currnet kind request")
		response.FailedWithError(c, apperrors.FailedRequest)
		return
	}
	ClusterID := c.Param("clusterId")
	namespace := c.Query("namespace")

	karmadaKubeClient := client.InClusterDynamicClientForKarmadaAPIServer()
	karmadaResourceList, err := listK8sResources(kind, namespace, karmadaKubeClient)
	if err != nil {
		log.Printf("failed request: %v", err)
		response.FailedWithError(c, apperrors.FailedRequest)
		return
	}

	credential, err := h.Adapter.GetKubeAccessInfo(ClusterID)
	if err != nil {
		log.Printf("failed request: %v", err)
		response.FailedWithError(c, apperrors.FailedRequest)
		return
	}
	memberClusterRestConfig := common.LoadRestConfigFromBearerToken(credential.APIServerURL, credential.BearerToken)
	clusterKubeClient, err := dynamic.NewForConfig(memberClusterRestConfig)
	if err != nil {
		log.Printf("failed request: %v", err)
		response.FailedWithError(c, apperrors.FailedRequest)
		return
	}
	clusterResourceList, err := listK8sResources(kind, namespace, clusterKubeClient)

	if err != nil {
		log.Printf("failed request: %v", err)
		response.FailedWithError(c, apperrors.FailedRequest)
		return
	} else {
		switch kind {
		case "deployment":
			var kResults []appsv1.Deployment
			var cResults []appsv1.Deployment
			var syncResourceList []SyncResource
			kResults, err = ConvertUnstructuredList[appsv1.Deployment](karmadaResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			cResults, err = ConvertUnstructuredList[appsv1.Deployment](clusterResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			syncResourceList = BuildSyncResourceList(ToPtrSlice(cResults), ToPtrSlice(kResults))
			if syncResourceList == nil {
				syncResourceList = make([]SyncResource, 0)
			}
			response.Success(c, syncResourceList)
		case "namespace":
			var kResults []corev1.Namespace
			var cResults []corev1.Namespace
			var syncResourceList []SyncResource
			kResults, err = ConvertUnstructuredList[corev1.Namespace](karmadaResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			cResults, err = ConvertUnstructuredList[corev1.Namespace](clusterResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			syncResourceList = BuildSyncResourceList(ToPtrSlice(cResults), ToPtrSlice(kResults))
			if syncResourceList == nil {
				syncResourceList = make([]SyncResource, 0)
			}
			response.Success(c, syncResourceList)
		case "statefulset":
			var kResults []appsv1.StatefulSet
			var cResults []appsv1.StatefulSet
			var syncResourceList []SyncResource
			kResults, err = ConvertUnstructuredList[appsv1.StatefulSet](karmadaResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			cResults, err = ConvertUnstructuredList[appsv1.StatefulSet](clusterResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			syncResourceList = BuildSyncResourceList(ToPtrSlice(cResults), ToPtrSlice(kResults))
			if syncResourceList == nil {
				syncResourceList = make([]SyncResource, 0)
			}
			response.Success(c, syncResourceList)
		case "daemonset":
			var kResults []appsv1.DaemonSet
			var cResults []appsv1.DaemonSet
			var syncResourceList []SyncResource
			kResults, err = ConvertUnstructuredList[appsv1.DaemonSet](karmadaResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			cResults, err = ConvertUnstructuredList[appsv1.DaemonSet](clusterResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			syncResourceList = BuildSyncResourceList(ToPtrSlice(cResults), ToPtrSlice(kResults))
			if syncResourceList == nil {
				syncResourceList = make([]SyncResource, 0)
			}
			response.Success(c, syncResourceList)
		case "job":
			var kResults []batch.Job
			var cResults []batch.Job
			var syncResourceList []SyncResource
			kResults, err = ConvertUnstructuredList[batch.Job](karmadaResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			cResults, err = ConvertUnstructuredList[batch.Job](clusterResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			syncResourceList = BuildSyncResourceList(ToPtrSlice(cResults), ToPtrSlice(kResults))
			if syncResourceList == nil {
				syncResourceList = make([]SyncResource, 0)
			}
			response.Success(c, syncResourceList)
		case "cronjob":
			var kResults []batch.CronJob
			var cResults []batch.CronJob
			var syncResourceList []SyncResource
			kResults, err = ConvertUnstructuredList[batch.CronJob](karmadaResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			cResults, err = ConvertUnstructuredList[batch.CronJob](clusterResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			syncResourceList = BuildSyncResourceList(ToPtrSlice(cResults), ToPtrSlice(kResults))
			if syncResourceList == nil {
				syncResourceList = make([]SyncResource, 0)
			}
			response.Success(c, syncResourceList)
		case "configmap":
			var kResults []corev1.ConfigMap
			var cResults []corev1.ConfigMap
			var syncResourceList []SyncResource
			kResults, err = ConvertUnstructuredList[corev1.ConfigMap](karmadaResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			cResults, err = ConvertUnstructuredList[corev1.ConfigMap](clusterResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			syncResourceList = BuildSyncResourceList(ToPtrSlice(cResults), ToPtrSlice(kResults))
			if syncResourceList == nil {
				syncResourceList = make([]SyncResource, 0)
			}
			response.Success(c, syncResourceList)
		case "secret":
			var kResults []corev1.Secret
			var cResults []corev1.Secret
			var syncResourceList []SyncResource
			kResults, err = ConvertUnstructuredList[corev1.Secret](karmadaResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			cResults, err = ConvertUnstructuredList[corev1.Secret](clusterResourceList)
			if err != nil {
				log.Printf("failed request: %v", err)
				response.FailedWithError(c, apperrors.FailedRequest)
				return
			}
			syncResourceList = BuildSyncResourceList(ToPtrSlice(cResults), ToPtrSlice(kResults))
			if syncResourceList == nil {
				syncResourceList = make([]SyncResource, 0)
			}
			response.Success(c, syncResourceList)
		}

	}
}

func (h Handler) HandlePostSync(c *gin.Context) {
	ClusterID := c.Param("clusterId")
	var SyncRequests SyncRequest
	err := json.NewDecoder(c.Request.Body).Decode(&SyncRequests)
	if err != nil {
		log.Printf("failed request: %v", err)
		response.FailedWithError(c, apperrors.RequestValueInvalid)
		return
	}

	var TotalResource = 0
	var FailResource = 0
	var SuccessResource = 0

	karmadaDynamicKubeClient := client.InClusterDynamicClientForKarmadaAPIServer()

	credential, err := h.Adapter.GetKubeAccessInfo(ClusterID)
	if err != nil {
		log.Printf("failed request: %v", err)
		response.FailedWithError(c, apperrors.FailedRequest)
		return
	}

	memberClusterRestConfig := common.LoadRestConfigFromBearerToken(credential.APIServerURL, credential.BearerToken)
	clusterKubeClient, err := dynamic.NewForConfig(memberClusterRestConfig)
	if err != nil {
		log.Printf("failed request: %v", err)
		response.FailedWithError(c, apperrors.FailedRequest)
		return
	}

	TotalResource = TotalResource + len(SyncRequests.CreateNamespaces)
	for _, createNs := range SyncRequests.CreateNamespaces {
		gvr1, _ := kindToGVR["namespace"]
		_, err := karmadaDynamicKubeClient.Resource(gvr1).Get(context.TODO(), createNs, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			srcObj, err := clusterKubeClient.Resource(gvr1).Get(context.TODO(), createNs, metav1.GetOptions{})
			if err != nil {
				// Namespace 생성 실패
				log.Printf("namespace create fail(%v): %v", createNs, err)
				FailResource++
				continue
			}
			objCopy := srcObj.DeepCopy()
			unstructured.RemoveNestedField(objCopy.Object, "metadata", "resourceVersion")
			unstructured.RemoveNestedField(objCopy.Object, "metadata", "uid")
			unstructured.RemoveNestedField(objCopy.Object, "metadata", "creationTimestamp")
			unstructured.RemoveNestedField(objCopy.Object, "metadata", "managedFields")
			unstructured.RemoveNestedField(objCopy.Object, "metadata", "generation")

			_, err = karmadaDynamicKubeClient.Resource(gvr1).Create(context.TODO(), objCopy, metav1.CreateOptions{})
			if err != nil {
				// Namespace 생성 실패
				log.Printf("namespace create fail(%v): %v", createNs, err)
				FailResource++
			} else {
				log.Printf("namespace create success(%v)", createNs)
				SuccessResource++
			}
		} else if err == nil {
			log.Printf("namespace create fail(%v): (already existed)", createNs)
			FailResource++
		} else {
			log.Printf("namespace create fail(%v) : %v", createNs, err)
			FailResource++
		}
	}

	for _, nsRes := range SyncRequests.Data {

		gvr1, _ := kindToGVR["namespace"]
		_, err := karmadaDynamicKubeClient.Resource(gvr1).Get(context.TODO(), nsRes.Namespace, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			//리소스를 만들수 없으므로 일괄 에러처리
			var failCount = 0
			for _, List := range nsRes.List {
				TotalResource = TotalResource + len(List.List)
				FailResource = FailResource + len(List.List)
				failCount++
			}
			log.Printf("resources create fail: namespace that doesn't exist(%v), fail count : %v", nsRes.Namespace, failCount)
			continue
		} else {
			for _, kindGroup := range nsRes.List {
				TotalResource = TotalResource + len(kindGroup.List)
				gvr, ok := kindToGVR[kindGroup.Kind]
				if !ok {
					// 지원하지 않는 kind
					log.Printf("unsupported kind(%v)", kindGroup.Kind)
					FailResource = FailResource + len(kindGroup.List)
					continue
				}

				for _, name := range kindGroup.List {
					// 원본 리소스 가져오기
					srcObj, err := clusterKubeClient.Resource(gvr).Namespace(nsRes.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
					if err != nil {
						// 리소스 가져오기 실패
						log.Printf("%v create fail(%v - %v) : %v", kindGroup.Kind, nsRes.Namespace, name, err)
						FailResource++
						continue
					}
					objCopy := srcObj.DeepCopy()
					unstructured.RemoveNestedField(objCopy.Object, "metadata", "resourceVersion")
					unstructured.RemoveNestedField(objCopy.Object, "metadata", "uid")
					unstructured.RemoveNestedField(objCopy.Object, "metadata", "creationTimestamp")
					unstructured.RemoveNestedField(objCopy.Object, "metadata", "managedFields")
					unstructured.RemoveNestedField(objCopy.Object, "metadata", "generation")

					if gvr.Resource == "jobs" {
						unstructured.RemoveNestedField(objCopy.Object, "metadata", "labels", "controller-uid")
						unstructured.RemoveNestedField(objCopy.Object, "metadata", "labels", "job-name")
						unstructured.RemoveNestedField(objCopy.Object, "metadata", "labels", "batch.kubernetes.io/controller-uid")
						unstructured.RemoveNestedField(objCopy.Object, "metadata", "labels", "batch.kubernetes.io/job-name")

						unstructured.RemoveNestedField(objCopy.Object, "spec", "selector")

						unstructured.RemoveNestedField(objCopy.Object, "spec", "template", "metadata", "labels", "controller-uid")
						unstructured.RemoveNestedField(objCopy.Object, "spec", "template", "metadata", "labels", "job-name")
						unstructured.RemoveNestedField(objCopy.Object, "spec", "template", "metadata", "labels", "batch.kubernetes.io/controller-uid")
						unstructured.RemoveNestedField(objCopy.Object, "spec", "template", "metadata", "labels", "batch.kubernetes.io/job-name")
					}

					_, err = karmadaDynamicKubeClient.Resource(gvr).Namespace(nsRes.Namespace).Create(context.TODO(), objCopy, metav1.CreateOptions{})
					if err != nil {
						if errors.IsAlreadyExists(err) {
							// Karmada에 이미 리소스 존재함
							log.Printf("%v create fail(%v - %v) : %v", kindGroup.Kind, nsRes.Namespace, name, err)
							FailResource++
						} else {
							// Karmada에 Sync 실패
							log.Printf("%v create fail(%v - %v) : %v", kindGroup.Kind, nsRes.Namespace, name, err)
							FailResource++
						}
					} else {
						// Karmada에 Sync 성공
						log.Printf("%v create success(%v - %v)", kindGroup.Kind, nsRes.Namespace, name)
						SuccessResource++

					}
				}
			}
		}
	}
	response.Success(c, SyncResponse{TotalResource: TotalResource, FailResource: FailResource, SuccessResource: SuccessResource})
}
