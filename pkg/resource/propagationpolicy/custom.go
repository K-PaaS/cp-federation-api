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
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	appcommon "github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/dashboard/pkg/common/helpers"
	"github.com/karmada-io/dashboard/pkg/common/types"
	"github.com/karmada-io/dashboard/pkg/dataselect"
	"github.com/karmada-io/dashboard/pkg/resource/common"
	"github.com/karmada-io/dashboard/pkg/resource/policy"
	"github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	karmadascheme "github.com/karmada-io/karmada/pkg/generated/clientset/versioned/scheme"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"log"
	"sigs.k8s.io/yaml"
)

// CustomPropagationPolicyList contains a list of propagation in the karmada control-plance.
type CustomPropagationPolicyList struct {
	ListMeta types.ListMeta `json:"listMeta"`

	// Unordered list of PropagationPolicies.
	PropagationPolicies []CustomPropagationPolicy `json:"propagationPolicies"`

	// List of non-critical errors, that occurred during resource retrieval.
	//Errors []error `json:"errors"`
}

// CustomPropagationPolicy contains information about a single propagation.
type CustomPropagationPolicy struct {
	Namespace          string   `json:"namespace"`
	Name               string   `json:"name"`
	UID                string   `json:"uid"`
	ConflictResolution string   `json:"conflictResolution"`
	RelatedClusters    []string `json:"relatedClusters"`
	// RelatedResources   []string `json:"relatedResources"`
}

// GetCustomPropagationPolicyList returns a list of all propagations in the karmada control-plance.
func GetCustomPropagationPolicyList(client karmadaclientset.Interface, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*CustomPropagationPolicyList, error) {
	log.Println("Getting list of namespaces")
	log.Println("nsQuery.ToRequestParam():", nsQuery.ToRequestParam())
	propagationpolicies, err := client.PolicyV1alpha1().PropagationPolicies(nsQuery.ToRequestParam()).List(context.TODO(), helpers.ListEverything)
	nonCriticalErrors, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return toCustomPropagationPolicyList(propagationpolicies.Items, nonCriticalErrors, dsQuery), nil
}

func toCustomPropagationPolicyList(propagationpolicies []v1alpha1.PropagationPolicy, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *CustomPropagationPolicyList {
	propagationpolicyList := &CustomPropagationPolicyList{
		PropagationPolicies: make([]CustomPropagationPolicy, 0),
		ListMeta:            types.ListMeta{TotalItems: len(propagationpolicies)},
	}
	propagationpolicyCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(propagationpolicies), dsQuery)
	propagationpolicies = fromCells(propagationpolicyCells)
	propagationpolicyList.ListMeta = types.ListMeta{TotalItems: filteredTotal}

	/*	verberClient, err := client.VerberClient(nil)
		if err != nil {
			panic(err)
		}
		for _, propagationpolicy := range propagationpolicies {
			relatedResources := policy.ExtractRelatedResources(verberClient, policy.PPWrapper{Policy: &propagationpolicy})
			pp := toCustomPropagationPolicy(&propagationpolicy)
			pp.RelatedResources = relatedResources
			propagationpolicyList.PropagationPolicies = append(propagationpolicyList.PropagationPolicies, pp)
		}*/

	for _, propagationpolicy := range propagationpolicies {
		pp := toCustomPropagationPolicy(&propagationpolicy)
		propagationpolicyList.PropagationPolicies = append(propagationpolicyList.PropagationPolicies, pp)
	}
	return propagationpolicyList
}

func toCustomPropagationPolicy(propagationpolicy *v1alpha1.PropagationPolicy) CustomPropagationPolicy {
	ObjectMeta := types.NewObjectMeta(propagationpolicy.ObjectMeta)
	relatedClusters, exists := policy.ExtractClusterNames(propagationpolicy.Spec.Placement.ClusterAffinity)
	if !exists {
		relatedClusters = make([]string, 0)
	}

	return CustomPropagationPolicy{
		Namespace:          ObjectMeta.Namespace,
		Name:               ObjectMeta.Name,
		UID:                string(ObjectMeta.UID),
		ConflictResolution: string(propagationpolicy.Spec.ConflictResolution),
		RelatedClusters:    relatedClusters,
	}
}

// GetCustomPropagationPolicyYaml gets propagationpolicy yaml.
func GetCustomPropagationPolicyYaml(ctx context.Context, client karmadaclientset.Interface, namespace, name string) (*v1.ResourceYaml, error) {
	pp, err := GetCustomPropagationPolicy(ctx, client, namespace, name)
	if err != nil {
		return nil, err
	}

	yaml, err := appcommon.EncodeToYAML(pp, karmadascheme.Scheme)
	if err != nil {
		return nil, err
	}

	return &v1.ResourceYaml{
		Namespace: pp.Namespace,
		Name:      pp.Name,
		UID:       pp.UID,
		Yaml:      yaml,
	}, nil
}

func GetCustomPropagationPolicy(ctx context.Context, client karmadaclientset.Interface, namespace, name string) (*v1alpha1.PropagationPolicy, error) {
	klog.Infof("Getting of PropagationPolicy %s/%s", namespace, name)
	pp, err := client.PolicyV1alpha1().PropagationPolicies(namespace).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get propagationPolicy")
		return nil, apperrors.ResourceError(err)
	}
	return pp, nil
}

func UpdateCustomPropagationPolicy(ctx context.Context, client karmadaclientset.Interface, namespace, name, updateYaml string) error {
	klog.Infof("Updating PropagationPolicy %s/%s", namespace, name)
	// 1. 존재하는 pp 인지 확인
	existingPP, err := GetCustomPropagationPolicy(ctx, client, namespace, name)
	if err != nil {
		return err
	}

	// 2. yaml -> pp unmarshal
	updatePP := v1alpha1.PropagationPolicy{}
	if err = yaml.Unmarshal([]byte(updateYaml), &updatePP); err != nil {
		klog.ErrorS(err, "Failed to unmarshal PropagationPolicy")
		return apperrors.InvalidYamlFormat
	}

	// 3. 네임스페이스, 이름, UID 검증
	if updatePP.Namespace != existingPP.Namespace ||
		updatePP.Name != existingPP.Name ||
		updatePP.UID != existingPP.UID {
		return apperrors.ResourceMismatch
	}

	// 4. ClusterAffinity.ClusterName nil, empty 체크
	updateClusterNames, exists := policy.ExtractClusterNames(updatePP.Spec.Placement.ClusterAffinity)
	if !exists {
		return apperrors.PolicyMissingTargetClusters
	}

	// ClusterAffinity.ClusterName 중복 제거
	updatePP.Spec.Placement.ClusterAffinity.ClusterNames = policy.RemoveDuplicates(updateClusterNames)

	_, err = client.PolicyV1alpha1().PropagationPolicies(namespace).Update(context.TODO(), &updatePP, metaV1.UpdateOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to update PropagationPolicy")
		return apperrors.ResourceError(err)
	}

	return nil
}

func DeleteCustomPropagationPolicy(ctx context.Context, client karmadaclientset.Interface, namespace, name string) error {
	klog.Infof("Deleting PropagationPolicy %s/%s", namespace, name)
	err := client.PolicyV1alpha1().PropagationPolicies(namespace).Delete(ctx, name, metaV1.DeleteOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to delete PropagationPolicy")
		return apperrors.ResourceError(err)
	}

	if retryErr := retry.OnError(
		retry.DefaultRetry,
		func(err error) bool {
			return !errors.IsNotFound(err)
		},
		func() error {
			_, getErr := client.PolicyV1alpha1().PropagationPolicies(namespace).Get(ctx, name, metaV1.GetOptions{})
			return getErr
		},
	); retryErr != nil {
		klog.ErrorS(retryErr, "PropagationPolicy deletion not confirmed")
		return apperrors.ResourceError(err)
	}

	return nil
}

func CreateCustomPropagationPolicy(ctx context.Context, client karmadaclientset.Interface, requestPP *v1.PostPropagationPolicyRequest) error {
	klog.Infof("Creating PropagationPolicy %s/%s", requestPP.Metadata.Namespace, requestPP.Metadata.Name)
	createPP := v1alpha1.PropagationPolicy{}

	if requestPP.Metadata.Namespace == "" {
		return apperrors.ResourceNamespaceRequired
	}

	// 1. build metadata
	objectMeta, err := policy.BuildPolicyMetadata(&requestPP.Metadata)
	if err != nil {
		return err
	}
	createPP.ObjectMeta = *objectMeta
	createPP.Spec.PreserveResourcesOnDeletion = requestPP.Metadata.PreserveResourcesOnDeletion

	// 2. build resourceSelector
	resourceSelectors, err := policy.BuildResourceSelectors(requestPP.ResourceSelectors, false)
	if err != nil {
		return err
	}
	createPP.Spec.ResourceSelectors = resourceSelectors

	// 3. build placement
	placement, err := policy.BuildPlacement(requestPP.Placement)
	if err != nil {
		return err
	}
	createPP.Spec.Placement = *placement

	// 4. create pp
	_, err = client.PolicyV1alpha1().PropagationPolicies(createPP.Namespace).Create(ctx, &createPP, metaV1.CreateOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to create PropagationPolicy")
		return apperrors.ResourceError(err)
	}

	return nil
}
