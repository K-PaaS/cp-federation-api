/*
Copyright 2024 The Karmada Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package overview

import (
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/metrics"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	"github.com/karmada-io/dashboard/cmd/api/app/routes/cluster"
	"k8s.io/klog/v2"
)

type Handler struct {
	ClusterHandler *cluster.Handler
}

func NewHandler(clusterHandler *cluster.Handler) *Handler {
	return &Handler{
		ClusterHandler: clusterHandler,
	}
}

func (h *Handler) HandleGetOverview(c *gin.Context) {
	karmadaInfo, err := GetControllerManagerInfo()
	if err != nil {
		klog.Errorf("Failed to get controller manager Info: %v", err)
		response.ServerError(c)
		return
	}

	metricsOpt, _ := metrics.GetClustersRealTimeUsage()
	memberClusterStatus, err := h.ClusterHandler.ListClusters(c, metricsOpt)
	if err != nil {
		klog.Errorf("Failed to get cluster list: %v", err)
		response.ServerError(c)
		return
	}

	data := CustomOverviewResponse{
		KarmadaInfo:           SetKarmadaInfo(karmadaInfo),
		ClusterResourceStatus: GetCustomClusterResourceStatus(),
		HostClusterStatus:     setHostClusterStatus(metricsOpt),
		MemberClusterStatus:   setMemberClusterStatus(memberClusterStatus.Clusters),
	}

	response.Success(c, data)
}
