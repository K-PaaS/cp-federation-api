package propagationpolicy

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	"github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/dashboard/pkg/resource/propagationpolicy"
)

func handleGetPropagationPolicyDetail(c *gin.Context) {
	karmadaClient := client.InClusterKarmadaClient()
	namespace := c.Param("namespace")
	name := c.Param("propagationPolicyName")
	result, err := propagationpolicy.GetPropagationPolicyDetail(karmadaClient, namespace, name)
	if err != nil {
		klog.ErrorS(err, "GetPropagationPolicyDetail failed")
		//response.FailRequest(c)
		return
	}
	response.Success(c, result)
}

func handleOrigDeletePropagationPolicy(c *gin.Context) {
	ctx := context.Context(c)
	propagationpolicyRequest := new(v1.DeletePropagationPolicyRequest)
	if err := c.ShouldBind(&propagationpolicyRequest); err != nil {
		common.Fail(c, err)
		return
	}
	var err error
	karmadaClient := client.InClusterKarmadaClient()
	if propagationpolicyRequest.IsClusterScope {
		err = karmadaClient.PolicyV1alpha1().ClusterPropagationPolicies().Delete(ctx, propagationpolicyRequest.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.ErrorS(err, "Failed to delete PropagationPolicy")
			common.Fail(c, err)
			return
		}
	} else {
		err = karmadaClient.PolicyV1alpha1().PropagationPolicies(propagationpolicyRequest.Namespace).Delete(ctx, propagationpolicyRequest.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.ErrorS(err, "Failed to delete PropagationPolicy")
			common.Fail(c, err)
			return
		}
		_ = retry.OnError(
			retry.DefaultRetry,
			func(err error) bool {
				return errors.IsNotFound(err)
			},
			func() error {
				_, getErr := karmadaClient.PolicyV1alpha1().PropagationPolicies(propagationpolicyRequest.Namespace).Get(ctx, propagationpolicyRequest.Name, metav1.GetOptions{})
				return getErr
			})
	}

	common.Success(c, "ok")
}

func handleOrigPutPropagationPolicy(c *gin.Context) {
	ctx := context.Context(c)
	propagationpolicyRequest := new(v1.PutPropagationPolicyRequest)
	if err := c.ShouldBind(&propagationpolicyRequest); err != nil {
		common.Fail(c, err)
		return
	}
	var err error
	karmadaClient := client.InClusterKarmadaClient()
	// todo check pp exist
	if propagationpolicyRequest.IsClusterScope {
		clusterpropagationPolicy := v1alpha1.ClusterPropagationPolicy{}
		if err = yaml.Unmarshal([]byte(propagationpolicyRequest.PropagationData), &clusterpropagationPolicy); err != nil {
			klog.ErrorS(err, "Failed to unmarshal ClusterPropagationPolicy")
			common.Fail(c, err)
			return
		}
		_, err = karmadaClient.PolicyV1alpha1().ClusterPropagationPolicies().Update(ctx, &clusterpropagationPolicy, metav1.UpdateOptions{})
	} else {
		propagationPolicy := v1alpha1.PropagationPolicy{}
		if err = yaml.Unmarshal([]byte(propagationpolicyRequest.PropagationData), &propagationPolicy); err != nil {
			klog.ErrorS(err, "Failed to unmarshal PropagationPolicy")
			common.Fail(c, err)
			return
		}
		var oldPropagationPolicy *v1alpha1.PropagationPolicy
		oldPropagationPolicy, err = karmadaClient.PolicyV1alpha1().PropagationPolicies(propagationpolicyRequest.Namespace).Get(ctx, propagationpolicyRequest.Name, metav1.GetOptions{})
		if err == nil {
			// only spec can be updated
			propagationPolicy.TypeMeta = oldPropagationPolicy.TypeMeta
			propagationPolicy.ObjectMeta = oldPropagationPolicy.ObjectMeta
			_, err = karmadaClient.PolicyV1alpha1().PropagationPolicies(propagationpolicyRequest.Namespace).Update(ctx, &propagationPolicy, metav1.UpdateOptions{})
		}
	}
	if err != nil {
		klog.ErrorS(err, "Failed to update PropagationPolicy")
		common.Fail(c, err)
		return
	}
	common.Success(c, "ok")
}
