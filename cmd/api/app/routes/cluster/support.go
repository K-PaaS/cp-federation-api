package cluster

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/domain"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/metrics"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	"github.com/karmada-io/dashboard/pkg/resource/cluster"
	"github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"time"
)

type RegisterResult struct {
	ClusterId string `json:"clusterId"`
	Name      string `json:"name"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

func (h *Handler) ListClusters(c *gin.Context, metricsOpt *metrics.ClusterUsage) (*cluster.CustomClusterList, error) {
	dataSelect := common.ParseDataSelectPathParameter(c)
	if err := common.IsValidProperties(dataSelect); err != nil {
		return nil, err
	}

	// Fetch clusters managed by the service
	managedClusters, err := h.Adapter.GetManagedClusters()
	if err != nil {
		return nil, apperrors.FailedRequest
	}
	managedClustersMap := make(map[string]string)
	for _, c := range managedClusters.Items {
		if c.IsFederated && c.FederatedClusterUID != "" {
			managedClustersMap[c.FederatedClusterUID] = c.ClusterID
		}
	}

	karmadaClient := client.InClusterKarmadaClient()
	result, err := cluster.GetCustomClusterList(karmadaClient, dataSelect, managedClustersMap, metricsOpt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h *Handler) RegisterMemberCluster(c *gin.Context, clusterID string, managedClusters []domain.Cluster) (*domain.FederationRequest, error) {
	clusterRequest := new(v1.PostClusterRequest)
	clusterRequest.ClusterID = clusterID

	// 1. 추가 가능한 클러스터인지 체크
	tc, err := checkRegisterableClusterByID(managedClusters, clusterRequest.ClusterID)
	if err != nil {
		return nil, err
	}

	clusterRequest.MemberClusterName = tc.Name
	clusterRequest.SyncMode = v1alpha1.Push

	// 2. 클러스터 인증 정보 조회 (endpoint,token)
	credential, err := h.Adapter.GetKubeAccessInfo(clusterRequest.ClusterID)
	if err != nil {
		klog.ErrorS(err, "GetKubeToken failed")
		return nil, apperrors.FailedToReadClusterInfo
	}

	// 3. karmadaClient 통한 멤버 클러스터 등록
	karmadaClient := client.InClusterKarmadaClient()
	memberClusterRestConfig := common.LoadRestConfigFromBearerToken(credential.APIServerURL, credential.BearerToken)
	restConfig, _, err := client.GetKarmadaConfig()
	if err != nil {
		klog.ErrorS(err, "Get restConfig failed")
		return nil, apperrors.ClusterLoadConfigFailed
	}
	opts := &pushModeOption{
		karmadaClient:           karmadaClient,
		clusterName:             clusterRequest.MemberClusterName,
		karmadaRestConfig:       restConfig,
		memberClusterRestConfig: memberClusterRestConfig,
	}
	if err := accessClusterInPushMode(opts); err != nil {
		klog.ErrorS(err, "accessClusterInPushMode failed")
		switch {
		case errors.Is(err, apperrors.ClusterAlreadyRegisteredInKarmada):
			return nil, apperrors.ClusterAlreadyRegisteredInKarmada
		default:
			return nil, apperrors.ClusterRegistrationFailed
		}
	}

	// 4. 등록된 멤버 클러스터 uid 조회
	clusterDetail, err := cluster.GetClusterDetail(karmadaClient, clusterRequest.MemberClusterName)
	if err != nil {
		klog.ErrorS(err, "GetClusterDetail failed")
		DeleteCluster(c, karmadaClient, clusterRequest.MemberClusterName)
		return nil, apperrors.ClusterRegistrationFailed
	}

	// 5. db 저장
	registerCluster := domain.FederationRequest{
		ClusterID:            clusterRequest.ClusterID,
		FederatedClusterUID:  string(clusterDetail.ObjectMeta.UID),
		FederatedClusterName: clusterRequest.MemberClusterName,
	}

	err = h.Adapter.RegisterFederatedCluster(&registerCluster)
	if err != nil {
		klog.ErrorS(err, "RegisterFederatedCluster failed")
		DeleteCluster(c, karmadaClient, clusterRequest.MemberClusterName)
		return nil, apperrors.ClusterRegistrationFailed
	}

	klog.Infof("accessClusterInPushMode success")
	return &registerCluster, nil
}

func DeleteCluster(c *gin.Context, karmadaClient karmadaclientset.Interface, clusterName string) error {
	ctx := context.Context(c)
	waitDuration := time.Second * 60
	err := karmadaClient.ClusterV1alpha1().Clusters().Delete(ctx, clusterName, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete cluster object. cluster name: %s, error: %v", clusterName, err)
		if apierrors.IsNotFound(err) {
			return apperrors.ClusterNotFoundInKarmada
		}
		return apperrors.FailedRequest
	}
	// make sure the given cluster object has been deleted
	err = wait.PollUntilContextTimeout(ctx, 1*time.Second, waitDuration, true, func(ctx context.Context) (done bool, err error) {
		_, err = karmadaClient.ClusterV1alpha1().Clusters().Get(ctx, clusterName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			klog.Errorf("Failed to get cluster %s. err: %v", clusterName, err)
			return false, err
		}
		klog.Infof("Waiting for the cluster object %s to be deleted", clusterName)
		return false, nil
	})
	if err != nil {
		klog.Errorf("Failed to delete cluster object. cluster name: %s, error: %v", clusterName, err)
		return apperrors.FailedRequest
	}

	return nil
}

func FindFederatedClusterByID(clusters []domain.Cluster, clusterID string) *domain.Cluster {
	for _, c := range clusters {
		if c.ClusterID == clusterID && c.IsFederated {
			return &c
		}
	}
	return nil
}

func FindRegisterableClusters(clusters []domain.Cluster) []domain.Cluster {
	result := make([]domain.Cluster, 0)
	for _, c := range clusters {
		if !c.IsFederated {
			result = append(result, c)
		}
	}
	return result
}

func checkRegisterableClusterByID(clusters []domain.Cluster, targetID string) (domain.Cluster, error) {
	for _, c := range clusters {
		if c.ClusterID != targetID {
			continue
		}
		if c.IsFederated {
			return c, apperrors.ClusterAlreadyRegistered
		}

		return c, nil
	}
	return domain.Cluster{}, apperrors.ClusterNotFound
}

/*

if clusterRequest.SyncMode == v1alpha1.Pull {
	memberClusterClient, err := commons.KubeClientSetFromBearerToken(credential.APIServerURL, credential.BearerToken)
	if err != nil {
		klog.ErrorS(err, "Generate kubeclient from memberClusterKubeconfig failed")
		response.Fail(c, err.Error())
		return
	}
	_, apiConfig, err := client.GetKarmadaConfig()
	if err != nil {
		klog.ErrorS(err, "Get apiConfig for karmada failed")
		response.Fail(c, err.Error())
		return
	}
	opts := &pullModeOption{
		karmadaClient:          karmadaClient,
		karmadaAgentCfg:        apiConfig,
		memberClusterNamespace: clusterRequest.MemberClusterNamespace,
		memberClusterClient:    memberClusterClient,
		memberClusterName:      clusterRequest.MemberClusterName,
		memberClusterEndpoint:  credential.APIServerURL,
	}
	if err = accessClusterInPullMode(opts); err != nil {
		klog.ErrorS(err, "accessClusterInPullMode failed")
		response.Fail(c, err.Error())
	} else {
		klog.Infof("accessClusterInPullMode success")
		response.Success(c, "ok")
	}
} else */

/*func handlePutCluster(c *gin.Context) {
	clusterRequest := new(v1.PutClusterRequest)
	name := c.Param("name")
	if err := c.ShouldBind(clusterRequest); err != nil {
		klog.ErrorS(err, "Could not read handlePutCluster request")
		common.Fail(c, err)
		return
	}
	karmadaClient := client.InClusterKarmadaClient()
	memberCluster, err := karmadaClient.ClusterV1alpha1().Clusters().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Get cluster failed")
		common.Fail(c, err)
		return
	}

	// assume that the frontend can fetch the whole labels and taints
	labels := make(map[string]string)
	if clusterRequest.Labels != nil {
		for _, labelItem := range *clusterRequest.Labels {
			labels[labelItem.Key] = labelItem.Value
		}
		memberCluster.Labels = labels
	}

	taints := make([]corev1.Taint, 0)
	if clusterRequest.Taints != nil {
		for _, taintItem := range *clusterRequest.Taints {
			taints = append(taints, corev1.Taint{
				Key:    taintItem.Key,
				Value:  taintItem.Value,
				Effect: taintItem.Effect,
			})
		}
		memberCluster.Spec.Taints = taints
	}

	_, err = karmadaClient.ClusterV1alpha1().Clusters().Update(context.TODO(), memberCluster, metav1.UpdateOptions{})
	if err != nil {
		klog.ErrorS(err, "Update cluster failed")
		common.Fail(c, err)
		return
	}
	common.Success(c, "ok")
}
*/
