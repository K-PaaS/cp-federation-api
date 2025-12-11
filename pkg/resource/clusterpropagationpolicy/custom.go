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
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/resource/policy"
	karmadascheme "github.com/karmada-io/karmada/pkg/generated/clientset/versioned/scheme"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	"github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"

	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/dashboard/pkg/common/helpers"
	"github.com/karmada-io/dashboard/pkg/common/types"
	"github.com/karmada-io/dashboard/pkg/dataselect"
)

// CustomClusterPropagationPolicyList contains a list of propagation in the karmada control-plane.
type CustomClusterPropagationPolicyList struct {
	ListMeta types.ListMeta `json:"listMeta"`

	// Unordered list of ClusterPropagationPolicies.
	ClusterPropagationPolicies []CustomClusterPropagationPolicy `json:"clusterPropagationPolicies"`

	// List of non-critical errors, that occurred during resource retrieval.
	//	Errors []error `json:"errors"`
}

// CustomClusterPropagationPolicy represents a cluster propagation policy.
type CustomClusterPropagationPolicy struct {
	Namespace          string   `json:"namespace,omitempty"`
	Name               string   `json:"name"`
	UID                string   `json:"uid"`
	ConflictResolution string   `json:"conflictResolution"`
	RelatedClusters    []string `json:"relatedClusters"`
	// RelatedResources   []string `json:"relatedResources"`
}

// GetCustomClusterPropagationPolicyList returns a list of all propagations in the karmada control-plance.
func GetCustomClusterPropagationPolicyList(client karmadaclientset.Interface, dsQuery *dataselect.DataSelectQuery) (*CustomClusterPropagationPolicyList, error) {
	clusterPropagationPolicies, err := client.PolicyV1alpha1().ClusterPropagationPolicies().List(context.TODO(), helpers.ListEverything)
	nonCriticalErrors, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return toCustomClusterPropagationPolicyList(clusterPropagationPolicies.Items, nonCriticalErrors, dsQuery), nil
}

func toCustomClusterPropagationPolicyList(clusterPropagationPolicies []v1alpha1.ClusterPropagationPolicy, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *CustomClusterPropagationPolicyList {
	propagationpolicyList := &CustomClusterPropagationPolicyList{
		ClusterPropagationPolicies: make([]CustomClusterPropagationPolicy, 0),
		ListMeta:                   types.ListMeta{TotalItems: len(clusterPropagationPolicies)},
	}
	clusterPropagationPolicyCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(clusterPropagationPolicies), dsQuery)
	clusterPropagationPolicies = fromCells(clusterPropagationPolicyCells)
	propagationpolicyList.ListMeta = types.ListMeta{TotalItems: filteredTotal}

	/*	verberClient, err := client.VerberClient(nil)
		if err != nil {
			panic(err)
		}
		for _, clusterPropagationPolicy := range clusterPropagationPolicies {
			relatedResources := policy.ExtractRelatedResources(verberClient, policy.CPPWrapper{Policy: &clusterPropagationPolicy})
			cpp := toCustomClusterPropagationPolicy(&clusterPropagationPolicy)
			cpp.RelatedResources = relatedResources
			propagationpolicyList.ClusterPropagationPolicies = append(propagationpolicyList.ClusterPropagationPolicies, cpp)
		}*/

	for _, clusterPropagationPolicy := range clusterPropagationPolicies {
		cpp := toCustomClusterPropagationPolicy(&clusterPropagationPolicy)
		propagationpolicyList.ClusterPropagationPolicies = append(propagationpolicyList.ClusterPropagationPolicies, cpp)
	}
	return propagationpolicyList
}

func toCustomClusterPropagationPolicy(propagationpolicy *v1alpha1.ClusterPropagationPolicy) CustomClusterPropagationPolicy {
	ObjectMeta := types.NewObjectMeta(propagationpolicy.ObjectMeta)
	relatedClusters, exists := policy.ExtractClusterNames(propagationpolicy.Spec.Placement.ClusterAffinity)
	if !exists {
		relatedClusters = make([]string, 0)
	}

	return CustomClusterPropagationPolicy{
		Namespace:          ObjectMeta.Namespace,
		Name:               ObjectMeta.Name,
		UID:                string(ObjectMeta.UID),
		ConflictResolution: string(propagationpolicy.Spec.ConflictResolution),
		RelatedClusters:    relatedClusters,
	}
}

func GetCustomClusterPropagationPolicy(ctx context.Context, client karmadaclientset.Interface, name string) (*v1alpha1.ClusterPropagationPolicy, error) {
	klog.Infof("Getting of ClusterPropagationPolicy %s", name)
	cpp, err := client.PolicyV1alpha1().ClusterPropagationPolicies().Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get ClusterPropagationPolicy")
		return nil, apperrors.ResourceError(err)
	}
	return cpp, nil
}

func GetCustomClusterPropagationPolicyYaml(ctx context.Context, client karmadaclientset.Interface, name string) (*v1.ResourceYaml, error) {
	cpp, err := GetCustomClusterPropagationPolicy(ctx, client, name)
	if err != nil {
		return nil, err
	}

	yaml, err := common.EncodeToYAML(cpp, karmadascheme.Scheme)
	if err != nil {
		return nil, err
	}

	return &v1.ResourceYaml{
		Name: cpp.Name,
		UID:  cpp.UID,
		Yaml: yaml,
	}, nil
}

func DeleteCustomClusterPropagationPolicy(ctx context.Context, client karmadaclientset.Interface, name string) error {
	klog.Infof("Deleting ClusterPropagationPolicy %s", name)
	err := client.PolicyV1alpha1().ClusterPropagationPolicies().Delete(ctx, name, metaV1.DeleteOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to delete ClusterPropagationPolicy")
		return apperrors.ResourceError(err)
	}

	if retryErr := retry.OnError(
		retry.DefaultRetry,
		func(err error) bool {
			return !errors.IsNotFound(err)
		},
		func() error {
			_, getErr := client.PolicyV1alpha1().ClusterPropagationPolicies().Get(ctx, name, metaV1.GetOptions{})
			return getErr
		},
	); retryErr != nil {
		klog.ErrorS(retryErr, "ClusterPropagationPolicy deletion not confirmed")
		return apperrors.ResourceError(err)
	}

	return nil
}

func UpdateCustomClusterPropagationPolicy(ctx context.Context, client karmadaclientset.Interface, name, updateYaml string) error {
	klog.Infof("Updating ClusterPropagationPolicy %s", name)
	// 1. 존재하는 cpp 인지 확인
	existingCPP, err := GetCustomClusterPropagationPolicy(ctx, client, name)
	if err != nil {
		return err
	}

	// 2. yaml -> cpp unmarshal
	updateCPP := v1alpha1.ClusterPropagationPolicy{}
	if err = yaml.Unmarshal([]byte(updateYaml), &updateCPP); err != nil {
		klog.ErrorS(err, "Failed to unmarshal ClusterPropagationPolicy")
		return apperrors.InvalidYamlFormat
	}

	// 3. 리소스 명, UID 검증
	if updateCPP.Name != existingCPP.Name ||
		updateCPP.UID != existingCPP.UID {
		return apperrors.ResourceMismatch
	}

	// 4. ClusterAffinity.ClusterName nil, empty 체크
	updateClusterNames, exists := policy.ExtractClusterNames(updateCPP.Spec.Placement.ClusterAffinity)
	if !exists {
		return apperrors.PolicyMissingTargetClusters
	}

	// ClusterAffinity.ClusterName 중복 제거
	updateCPP.Spec.Placement.ClusterAffinity.ClusterNames = policy.RemoveDuplicates(updateClusterNames)
	_, err = client.PolicyV1alpha1().ClusterPropagationPolicies().Update(ctx, &updateCPP, metaV1.UpdateOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to update ClusterPropagationPolicy")
		return apperrors.ResourceError(err)
	}

	return nil
}

func CreateCustomClusterPropagationPolicy(ctx context.Context, client karmadaclientset.Interface, requestCPP *v1.PostPropagationPolicyRequest) error {
	klog.Infof("Creating ClusterPropagationPolicy %s", requestCPP.Metadata.Name)
	createCPP := v1alpha1.ClusterPropagationPolicy{}

	// 1. build metadata
	objectMeta, err := policy.BuildPolicyMetadata(&requestCPP.Metadata)
	if err != nil {
		return err
	}
	createCPP.ObjectMeta = *objectMeta
	createCPP.Spec.PreserveResourcesOnDeletion = requestCPP.Metadata.PreserveResourcesOnDeletion

	// 2. build resourceSelector
	resourceSelectors, err := policy.BuildResourceSelectors(requestCPP.ResourceSelectors, true)
	if err != nil {
		return err
	}
	createCPP.Spec.ResourceSelectors = resourceSelectors

	// 3. build placement
	placement, err := policy.BuildPlacement(requestCPP.Placement)
	if err != nil {
		return err
	}
	createCPP.Spec.Placement = *placement

	// 4. create cpp
	_, err = client.PolicyV1alpha1().ClusterPropagationPolicies().Create(ctx, &createCPP, metaV1.CreateOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to create ClusterPropagationPolicy")
		return apperrors.ResourceError(err)
	}
	return nil
}
