package federation

import (
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/adapter/auth"
	"github.com/karmada-io/dashboard/cmd/api/app/domain"
)

type MockAdapter struct {
	Auth auth.AuthAdapter
}

func NewMockAdapter(authAdapter auth.AuthAdapter) FederationAdapter {
	return &InternalAdapter{
		Auth: authAdapter,
	}
}

func (m *MockAdapter) GetKubeAccessInfo(clusterID string) (*domain.ClusterCredential, error) {
	//TODO implement me

	return &domain.ClusterCredential{
		ClusterID:    clusterID,
		APIServerURL: "https://mock-cluster-api-server",
		BearerToken:  "mock-fake-token-1234",
	}, nil
}

func (m *MockAdapter) GetManagedClusters(c *gin.Context) (*domain.ClusterList, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockAdapter) RegisterFederatedCluster(federationRequest *domain.FederationRequest) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockAdapter) UnregisterFederatedCluster(clusterID string) error {
	//TODO implement me
	panic("implement me")
}
