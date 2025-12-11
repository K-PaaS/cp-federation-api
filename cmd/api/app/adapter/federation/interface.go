package federation

import (
	"github.com/karmada-io/dashboard/cmd/api/app/domain"
)

type FederationAdapter interface {
	GetKubeAccessInfo(clusterID string) (*domain.ClusterCredential, error)
	GetManagedClusters() (*domain.ClusterList, error)
	RegisterFederatedCluster(federationRequest *domain.FederationRequest) error
	UnregisterFederatedCluster(clusterID string) error
}
