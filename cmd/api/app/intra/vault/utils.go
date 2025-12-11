package vault

import (
	"fmt"
	"github.com/karmada-io/dashboard/cmd/api/app/domain"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/intra"
	"log/slog"
)

func GetClusterCredential(clusterID string) (*domain.ClusterCredential, error) {
	credential := &domain.ClusterCredential{
		ClusterID: clusterID,
	}
	if err := readClusterCredential(credential); err != nil {
		return nil, err
	}

	return credential, nil
}

func readClusterCredential(credential *domain.ClusterCredential) error {
	path := fmt.Sprintf("%v/%v", intra.Env.VaultClusterPath, credential.ClusterID)
	data, err := read(path)
	if err != nil {
		slog.Error("getClusterDetails", "err", err)
		return apperrors.FailedToReadClusterInfo
	}

	credential.APIServerURL = data["clusterApiUrl"].(string)
	credential.BearerToken = data["clusterToken"].(string)
	return nil
}
