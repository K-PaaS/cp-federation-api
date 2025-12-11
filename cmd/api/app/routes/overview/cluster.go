package overview

import (
	"github.com/karmada-io/dashboard/cmd/api/app/metrics"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/pkg/resource/cluster"
)

// CustomOverviewResponse custom
type CustomOverviewResponse struct {
	KarmadaInfo           *KarmadaInfo              `json:"karmadaInfo"`
	ClusterResourceStatus *v1.ClusterResourceStatus `json:"clusterResourceStatus"`
	HostClusterStatus     metrics.Status            `json:"hostClusterStatus"`
	MemberClusterStatus   []metrics.Status          `json:"memberClusterStatus"`
}

func setHostClusterStatus(metricsOpt *metrics.ClusterUsage) metrics.Status {
	if metricsOpt != nil {
		metricsOpt.HostClusterStatus.Status = cluster.GetConditionStatus(metricsOpt.HostClusterStatus.Status)
		return metricsOpt.HostClusterStatus
	}
	return metrics.Status{
		NodeSummary:   &v1.NodeSummary{},
		RealTimeUsage: v1.InitUsage(),
		RequestUsage:  v1.InitUsage(),
	}
}

func setMemberClusterStatus(clusters []cluster.CustomCluster) []metrics.Status {
	memberStatus := make([]metrics.Status, 0)
	for _, c := range clusters {
		memberStatus = append(memberStatus, toMemberClusterStatus(&c))
	}
	return memberStatus
}

func toMemberClusterStatus(cluster *cluster.CustomCluster) metrics.Status {
	return metrics.Status{
		ClusterId:     cluster.ClusterId,
		Name:          cluster.Name,
		UID:           cluster.UID,
		Status:        cluster.Status,
		NodeSummary:   cluster.NodeSummary,
		RealTimeUsage: cluster.RealTimeUsage,
		RequestUsage:  cluster.RequestUsage,
	}
}
