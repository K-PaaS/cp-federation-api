package federation

import (
	"github.com/karmada-io/dashboard/cmd/api/app/adapter/auth"
	"github.com/karmada-io/dashboard/cmd/api/app/domain"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/intra"
	"github.com/karmada-io/dashboard/cmd/api/app/intra/vault"
	"k8s.io/klog/v2"
	"net/http"
)

type InternalAdapter struct {
	Auth auth.AuthAdapter
}

func NewInternalAdapter(authAdapter auth.AuthAdapter) FederationAdapter {
	return &InternalAdapter{
		Auth: authAdapter,
	}
}

func (v InternalAdapter) GetKubeAccessInfo(clusterID string) (*domain.ClusterCredential, error) {
	credential, err := vault.GetClusterCredential(clusterID)
	if err != nil {
		return nil, err
	}
	return credential, nil
}

func (v InternalAdapter) GetManagedClusters() (*domain.ClusterList, error) {
	resp, err := intra.ApiCallManagedClusters()
	if err != nil {
		return nil, err
	}
	return &domain.ClusterList{
		Items: resp.Items,
	}, nil
}

func (v InternalAdapter) RegisterFederatedCluster(federationRequest *domain.FederationRequest) error {
	var resp intra.ClusterResponse
	err := intra.ApiCall(intra.CommonApi, http.MethodPost, intra.Env.CreateFederatedClusterUrl, federationRequest, &resp)
	if err != nil {
		klog.ErrorS(err, "RegisterFederatedCluster failed")
		return apperrors.FailedRequest
	}
	if resp.ResultCode == intra.ResultStatusFail {
		klog.Error(resp.ResultMessage)
		return apperrors.FailedRequest
	}

	return nil
}

func (v InternalAdapter) UnregisterFederatedCluster(clusterID string) error {
	var resp intra.ClusterResponse
	path := intra.Env.DeleteFederatedClusterUrl + clusterID
	err := intra.ApiCall(intra.CommonApi, http.MethodDelete, path, nil, &resp)
	if err != nil {
		klog.ErrorS(err, "ApiCall failed")
		return apperrors.FailedRequest
	}
	if resp.ResultCode == intra.ResultStatusFail {
		klog.Error(resp.ResultMessage)
		return apperrors.FailedRequest
	}
	return nil
}
