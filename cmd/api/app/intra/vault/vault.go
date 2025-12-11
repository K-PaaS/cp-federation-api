package vault

import (
	"context"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/karmada-io/dashboard/cmd/api/app/intra"
	"log/slog"
	"time"
)

func read(path string) (map[string]interface{}, error) {
	slog.Info("read value", "path", path)
	ctx := context.Background()
	//get vault client
	client, err := getVaultClient()
	if err != nil {
		return nil, err
	}

	// read a secret
	resp, err := client.Read(ctx, path)
	if err != nil {
		return nil, err
	}

	return resp.Data["data"].(map[string]interface{}), nil
}

func getVaultClient() (*vault.Client, error) {
	ctx := context.Background()
	// prepare a client with the given base address
	client, err := vault.New(
		vault.WithAddress(intra.Env.VaultUrl),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, err
	}

	// authenticate using approle
	resp, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   intra.Env.VaultRoleId,
			SecretId: intra.Env.VaultSecretId,
		},
		vault.WithMountPath(""),
	)
	if err != nil {
		return nil, err
	}

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		return nil, err
	}

	return client, nil
}
