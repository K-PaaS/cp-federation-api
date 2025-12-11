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

package cluster

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/adapter/federation"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/localize"
	errmsg "github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/pkg/client"
	"github.com/karmada-io/dashboard/pkg/resource/cluster"
	"k8s.io/klog/v2"
	"net/http"
	"sync"
)

type Handler struct {
	Adapter federation.FederationAdapter
}

func NewHandler(adapter federation.FederationAdapter) *Handler {
	return &Handler{
		Adapter: adapter,
	}
}

const (
	RegisteredSuccessfully = "CLUSTER_REGISTERED_SUCCESSFULLY"
	DeletionCompleted      = "CLUSTER_DELETION_COMPLETED"
	clustersKey            = "clusters"
)

func (h *Handler) HandleListClusters(c *gin.Context) {
	result, err := h.ListClusters(c, nil)
	if err != nil {
		klog.ErrorS(err, "GetCustomClusterList failed")
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) HandleGetClusterYaml(c *gin.Context) {
	clusterId := c.Param("clusterId")
	// Fetch clusters managed by the service
	managedClusters, err := h.Adapter.GetManagedClusters()
	if err != nil {
		response.FailedWithError(c, err)
		return
	}

	target := FindFederatedClusterByID(managedClusters.Items, clusterId)
	if target == nil {
		response.FailedWithError(c, apperrors.ClusterNotFound)
		return
	}

	karmadaClient := client.InClusterKarmadaClient()
	result, err := cluster.GetCustomClusterYaml(karmadaClient, target.ClusterID, target.FederatedClusterName)
	if err != nil {
		klog.ErrorS(err, "GetCustomClusterYaml failed")
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) HandleListRegisterableClusters(c *gin.Context) {
	// Fetch clusters managed by the service
	managedClusters, err := h.Adapter.GetManagedClusters()
	if err != nil {
		response.FailedWithError(c, err)
		return
	}
	result := FindRegisterableClusters(managedClusters.Items)
	response.Success(c, gin.H{clustersKey: result})
}

func (h *Handler) HandlePostCluster(c *gin.Context) {
	clusterRequest := new(v1.PostClusterRequest)
	if err := c.ShouldBind(clusterRequest); err != nil {
		klog.ErrorS(err, "Could not read cluster request value")
		response.FailedWithError(c, apperrors.RequestValueInvalid)
		return
	}

	// Fetch clusters managed by the service
	managedClusters, err := h.Adapter.GetManagedClusters()
	if err != nil {
		response.FailedWithError(c, err)
		return
	}

	var wg sync.WaitGroup
	resultChan := make(chan RegisterResult, len(clusterRequest.ClusterIDs))
	for _, cid := range clusterRequest.ClusterIDs {
		wg.Add(1)
		go func(targetClusterId string) {
			defer wg.Done()
			result := RegisterResult{ClusterId: targetClusterId}
			register, err := h.RegisterMemberCluster(c, targetClusterId, managedClusters.Items)
			if err != nil {
				result.Code = http.StatusInternalServerError
				result.Message = localize.GetLocalizeMessage(c, errmsg.RequestFailed)
				var httpErr *apperrors.HttpError
				if errors.As(err, &httpErr) {
					result.Code = httpErr.Code
					result.Message = localize.GetLocalizeMessage(c, httpErr.Msg)
				}
			} else {
				result.Code = http.StatusCreated
				result.Name = register.FederatedClusterName
				result.Message = localize.GetLocalizeMessage(c, RegisteredSuccessfully)
			}
			resultChan <- result
		}(cid)
	}
	wg.Wait()
	close(resultChan)

	var results []RegisterResult
	for r := range resultChan {
		results = append(results, r)
	}

	response.Success(c, gin.H{clustersKey: results})
}

func (h *Handler) HandleDeleteCluster(c *gin.Context) {
	clusterRequest := new(v1.DeleteClusterRequest)
	if err := c.ShouldBindUri(&clusterRequest); err != nil {
		response.FailedWithError(c, apperrors.RequestValueInvalid)
		return
	}

	// Fetch clusters managed by the service
	managedClusters, err := h.Adapter.GetManagedClusters()
	if err != nil {
		response.FailedWithError(c, err)
		return
	}

	target := FindFederatedClusterByID(managedClusters.Items, clusterRequest.ClusterID)
	if target == nil {
		response.FailedWithError(c, apperrors.ClusterNotFound)
		return
	}

	// Delete member from karmada
	karmadaClient := client.InClusterKarmadaClient()
	err = DeleteCluster(c, karmadaClient, target.FederatedClusterName)
	if err != nil {
		response.FailedWithError(c, err)
		return
	}
	// Unregister FederatedCluster Mapping
	err = h.Adapter.UnregisterFederatedCluster(clusterRequest.ClusterID)
	if err != nil {
		klog.ErrorS(err, "UnregisterFederatedCluster failed")
		response.FailedWithError(c, err)
		return
	}

	response.SuccessWithMessage(c, DeletionCompleted)
}
