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

package clusterpropagationpolicy

import (
	"context"
	"github.com/gin-gonic/gin"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	"github.com/karmada-io/dashboard/pkg/resource/clusterpropagationpolicy"
	"k8s.io/klog/v2"
)

func HandleGetClusterPropagationPolicyList(c *gin.Context) {
	dataSelect := common.ParseDataSelectPathParameter(c)
	if err := common.IsValidProperties(dataSelect); err != nil {
		response.FailedWithError(c, err)
		return
	}

	karmadaClient := client.InClusterKarmadaClient()
	clusterPropagationList, err := clusterpropagationpolicy.GetCustomClusterPropagationPolicyList(karmadaClient, dataSelect)
	if err != nil {
		klog.ErrorS(err, "Failed to GetCustomClusterPropagationPolicyList")
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, clusterPropagationList)
}

func HandleGetClusterPropagationPolicyYaml(c *gin.Context) {
	ctx := context.Context(c)
	karmadaClient := client.InClusterKarmadaClient()
	name := c.Param("clusterPropagationPolicyName")
	result, err := clusterpropagationpolicy.GetCustomClusterPropagationPolicyYaml(ctx, karmadaClient, name)
	if err != nil {
		klog.ErrorS(err, "GetCustomClusterPropagationPolicyYaml failed")
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, result)
}

func HandlePostClusterPropagationPolicy(c *gin.Context) {
	ctx := context.Context(c)
	createCPP := new(v1.PostPropagationPolicyRequest)
	if err := c.ShouldBind(&createCPP); err != nil {
		response.FailedWithError(c, apperrors.RequestValueInvalid)
		return
	}

	karmadaClient := client.InClusterKarmadaClient()
	err := clusterpropagationpolicy.CreateCustomClusterPropagationPolicy(ctx, karmadaClient, createCPP)
	if err != nil {
		klog.ErrorS(err, "Failed to create ClusterPropagationPolicy")
		response.FailedWithError(c, err)
		return
	}
	response.Created(c)
}

func HandleDeleteClusterPropagationPolicy(c *gin.Context) {
	ctx := context.Context(c)
	karmadaClient := client.InClusterKarmadaClient()
	name := c.Param("clusterPropagationPolicyName")
	err := clusterpropagationpolicy.DeleteCustomClusterPropagationPolicy(ctx, karmadaClient, name)
	if err != nil {
		klog.ErrorS(err, "Failed to delete ClusterPropagationPolicy")
		response.FailedWithError(c, err)
		return
	}
	response.SuccessWithMessage(c, msgkey.ResourceDeletionCompleted)
}

func HandlePutClusterPropagationPolicy(c *gin.Context) {
	updateCPP := new(v1.PutPropagationPolicyRequest)
	if err := c.ShouldBind(&updateCPP); err != nil {
		response.FailedWithError(c, apperrors.RequestValueInvalid)
		return
	}

	ctx := context.Context(c)
	karmadaClient := client.InClusterKarmadaClient()
	name := c.Param("clusterPropagationPolicyName")

	err := clusterpropagationpolicy.UpdateCustomClusterPropagationPolicy(ctx, karmadaClient, name, updateCPP.PropagationData)
	if err != nil {
		klog.ErrorS(err, "Failed to update ClusterPropagationPolicy")
		response.FailedWithError(c, err)
		return
	}

	response.SuccessWithMessage(c, msgkey.ResourceUpdateSuccess)
}
