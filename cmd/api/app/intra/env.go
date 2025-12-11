package intra

import (
	"github.com/spf13/viper"
	"log/slog"
	"os"
)

var Env *envConfigs

func init() {
	Env = loadEnvVariables()
}

type envConfigs struct {
	JwtSecret                  string   `mapstructure:"JwtSecret"`
	VaultUrl                   string   `mapstructure:"VaultUrl"`
	VaultRoleId                string   `mapstructure:"VaultRoleId"`
	VaultSecretId              string   `mapstructure:"VaultSecretId"`
	VaultClusterPath           string   `mapstructure:"VaultClusterPath"`
	CommonApiUrl               string   `mapstructure:"CommonApiUrl"`
	CommonApiUserName          string   `mapstructure:"CommonApiUserName"`
	CommonApiUserPassword      string   `mapstructure:"CommonApiUserPassword"`
	GetFederatedClusterListUrl string   `mapstructure:"GetFederatedClusterListUrl"`
	CreateFederatedClusterUrl  string   `mapstructure:"CreateFederatedClusterUrl"`
	DeleteFederatedClusterUrl  string   `mapstructure:"DeleteFederatedClusterUrl"`
	IsSuperAdminCheckUrl       string   `mapstructure:"IsSuperAdminCheckUrl"`
	FilterNamespaces           []string `mapstructure:"FilterNamespaces"`
}

func loadEnvVariables() (config *envConfigs) {
	cwd, _ := os.Getwd()
	slog.Info("cwd", "intra/env.go", cwd)
	v := viper.New()
	v.AddConfigPath("app/intra")
	v.SetConfigName("config")
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		slog.Error("Error reading env file", "err", err)
	}
	if err := v.Unmarshal(&config); err != nil {
		slog.Error(err.Error())
	}
	return
}
