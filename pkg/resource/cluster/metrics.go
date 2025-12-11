package cluster

import (
	"github.com/karmada-io/dashboard/cmd/api/app/metrics"
	"k8s.io/klog/v2"
)

func setClusterRealTimeUsage(clusters []CustomCluster, metricsOpt *metrics.ClusterUsage) {
	var metricsData *metrics.ClusterUsage

	if metricsOpt != nil {
		metricsData = metricsOpt
	} else {
		metricsData, _ = metrics.GetClustersRealTimeUsage()
	}

	if metricsData == nil {
		klog.Warning("Cluster usage metrics are nil")
		return
	}

	memberStatusMap := make(map[string]metrics.Status)
	for _, status := range metricsData.MemberClusterStatus {
		memberStatusMap[status.ClusterId] = status
	}

	// Match and update each cluster's RealTimeUsage
	for i := range clusters {
		cluster := &clusters[i]
		if status, ok := memberStatusMap[cluster.ClusterId]; ok {
			cluster.RealTimeUsage.CPU = status.RealTimeUsage.CPU
			cluster.RealTimeUsage.Memory = status.RealTimeUsage.Memory
		}
	}
}
