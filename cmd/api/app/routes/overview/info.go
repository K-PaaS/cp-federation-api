package overview

import v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"

type KarmadaInfo struct {
	Version    string `json:"version"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
}

func SetKarmadaInfo(karmadaInfo *v1.KarmadaInfo) *KarmadaInfo {
	return &KarmadaInfo{
		Version:    karmadaInfo.Version.GitVersion,
		Status:     karmadaInfo.Status,
		CreateTime: karmadaInfo.CreateTime.Format("2006-01-02 15:04:05"),
	}
}
