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

package namespace

import (
	"context"
	"github.com/gin-gonic/gin"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	internalCommon "github.com/karmada-io/dashboard/pkg/resource/common"
	"github.com/karmada-io/dashboard/pkg/resource/event"
	ns "github.com/karmada-io/dashboard/pkg/resource/namespace"
	"k8s.io/klog/v2"
)

func HandleCreateNamespace(c *gin.Context) {
	createNamespaceRequest := new(v1.CreateNamespaceRequest)
	if err := c.ShouldBind(&createNamespaceRequest); err != nil {
		response.FailedWithError(c, apperrors.RequestValueInvalid)
		return
	}
	/*
		if internalCommon.IsFilterNamespaceCheck(createNamespaceRequest.Name) {
			klog.ErrorS(nil, "Can't use prefix"+createNamespaceRequest.Name)
			response.FailedWithError(c, apperrors.NotAllowedNamespace)
			return
		}
	*/

	k8sClient := client.InClusterClientForKarmadaAPIServer()
	err := ns.CreateNamespace(createNamespaceRequest, k8sClient)
	if err != nil {
		klog.ErrorS(err, "Failed to create namespace")
		response.FailedWithError(c, err)
		return
	}
	response.Created(c)
}
func HandleGetNamespaceList(c *gin.Context) {
	dataSelect := common.ParseDataSelectPathParameter(c)
	if err := common.IsValidProperties(dataSelect); err != nil {
		response.FailedWithError(c, err)
		return
	}
	k8sClient := client.InClusterClientForKarmadaAPIServer()
	result, err := ns.GetNamespaceList(k8sClient, dataSelect)
	if err != nil {
		klog.ErrorS(err, "Failed to get namespace list")
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, result)
}

func HandleDeleteNamespace(c *gin.Context) {
	ctx := context.Context(c)
	name := c.Param("namespaceName")

	if internalCommon.IsFilterNamespaceCheck(name) {
		klog.ErrorS(nil, "Can't use prefix"+name)
		response.FailedWithError(c, apperrors.NotAllowedNamespace)
		return
	}
	k8sClient := client.InClusterClientForKarmadaAPIServer()
	err := ns.DeleteNamespace(ctx, k8sClient, name)
	if err != nil {
		klog.ErrorS(err, "Failed to delete namespace")
		response.FailedWithError(c, err)
		return
	}
	response.SuccessWithMessage(c, msgkey.ResourceDeletionCompleted)
}

func HandleGetNamespaceYaml(c *gin.Context) {
	name := c.Param("namespaceName")
	/*
		if internalCommon.IsFilterNamespaceCheck(name) {
			klog.ErrorS(nil, "Can't use prefix"+name)
			response.FailedWithError(c, apperrors.NotAllowedNamespace)
			return
		}
	*/
	k8sClient := client.InClusterClientForKarmadaAPIServer()
	result, err := ns.GetNamespaceYaml(k8sClient, name)
	if err != nil {
		klog.ErrorS(err, "Failed to get namespace yaml")
		response.FailedWithError(c, err)
		return
	}
	response.Success(c, result)
}
func handleGetNamespaceEvents(c *gin.Context) {
	k8sClient := client.InClusterClientForKarmadaAPIServer()
	name := c.Param("name")
	dataSelect := common.ParseDataSelectPathParameter(c)
	result, err := event.GetNamespaceEvents(k8sClient, dataSelect, name)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Success(c, result)
}
