package metrics

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
	NatsUsername    string `mapstructure:"NatsUsername"`
	NatsPassword    string `mapstructure:"NatsPassword"`
	NatsUrl         string `mapstructure:"NatsUrl"`
	NatsBucketName  string `mapstructure:"NatsBucketName"`
	NatsSubjectName string `mapstructure:"NatsSubjectName"`
}

func loadEnvVariables() (config *envConfigs) {
	cwd, _ := os.Getwd()
	slog.Info("cwd", "metrics/env.go", cwd)
	v := viper.New()
	v.AddConfigPath("app/metrics")
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
