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

package propagationpolicy

import (
	"context"
	"github.com/gin-gonic/gin"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	"github.com/karmada-io/dashboard/pkg/resource/propagationpolicy"
	"k8s.io/klog/v2"
)

func HandleGetPropagationPolicyList(c *gin.Context) {
	dataSelect := common.ParseDataSelectPathParameter(c)
	if err := common.IsValidProperties(dataSelect); err != nil {
		response.FailedWithError(c, err)
		return
	}

	karmadaClient := client.InClusterKarmadaClient()
	namespace := common.ParseNamespaceQuery(c)
	propagationList, err := propagationpolicy.GetCustomPropagationPolicyList(karmadaClient, namespace, dataSelect)
	if err != nil {
		klog.ErrorS(err, "Failed to GetCustomPropagationPolicyList")
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, propagationList)
}

func HandleGetPropagationPolicyYaml(c *gin.Context) {
	ctx := context.Context(c)
	karmadaClient := client.InClusterKarmadaClient()
	namespace := c.Param("namespace")
	name := c.Param("propagationPolicyName")
	result, err := propagationpolicy.GetCustomPropagationPolicyYaml(ctx, karmadaClient, namespace, name)
	if err != nil {
		klog.ErrorS(err, "GetCustomPropagationPolicyYaml failed")
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, result)
}

func HandlePostPropagationPolicy(c *gin.Context) {
	ctx := context.Context(c)
	createPP := new(v1.PostPropagationPolicyRequest)
	if err := c.ShouldBind(&createPP); err != nil {
		response.FailedWithError(c, apperrors.RequestValueInvalid)
		return
	}

	karmadaClient := client.InClusterKarmadaClient()
	err := propagationpolicy.CreateCustomPropagationPolicy(ctx, karmadaClient, createPP)
	if err != nil {
		klog.ErrorS(err, "Failed to create PropagationPolicy")
		response.FailedWithError(c, err)
		return
	}
	response.Created(c)
}

func HandlePutPropagationPolicy(c *gin.Context) {
	updatePP := new(v1.PutPropagationPolicyRequest)
	if err := c.ShouldBind(&updatePP); err != nil {
		response.FailedWithError(c, apperrors.RequestValueInvalid)
		return
	}

	ctx := context.Context(c)
	karmadaClient := client.InClusterKarmadaClient()
	namespace := c.Param("namespace")
	name := c.Param("propagationPolicyName")

	err := propagationpolicy.UpdateCustomPropagationPolicy(ctx, karmadaClient, namespace, name, updatePP.PropagationData)
	if err != nil {
		klog.ErrorS(err, "Failed to update PropagationPolicy")
		response.FailedWithError(c, err)
		return
	}

	response.SuccessWithMessage(c, msgkey.ResourceUpdateSuccess)
}

func HandleDeletePropagationPolicy(c *gin.Context) {
	ctx := context.Context(c)
	karmadaClient := client.InClusterKarmadaClient()
	namespace := c.Param("namespace")
	name := c.Param("propagationPolicyName")

	err := propagationpolicy.DeleteCustomPropagationPolicy(ctx, karmadaClient, namespace, name)
	if err != nil {
		klog.ErrorS(err, "Failed to delete PropagationPolicy")
		response.FailedWithError(c, err)
		return
	}
	response.SuccessWithMessage(c, msgkey.ResourceDeletionCompleted)
}
