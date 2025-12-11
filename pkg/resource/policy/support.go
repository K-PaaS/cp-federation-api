package policy

import (
	"fmt"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	appcommon "github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"strings"
)

type ClusterRelation int

const (
	RelationDisjoint ClusterRelation = iota
	RelationPartial
	RelationEqual
	AnnotationKeyAccessLevel = "cp.internal.federation.api/access-level"
	AccessLevelReadOnly      = "readonly"
	AccessLevelFull          = "full"
)

type HasClusterNames interface {
	GetClusterNames() ([]string, bool)
}
type HasResourceSelectors interface {
	GetResourceSelectors() []v1alpha1.ResourceSelector
}

type PPWrapper struct {
	Policy *v1alpha1.PropagationPolicy
}
type CPPWrapper struct {
	Policy *v1alpha1.ClusterPropagationPolicy
}

func (w PPWrapper) GetClusterNames() ([]string, bool) {
	return ExtractClusterNames(w.Policy.Spec.Placement.ClusterAffinity)
}

func (w CPPWrapper) GetClusterNames() ([]string, bool) {
	return ExtractClusterNames(w.Policy.Spec.Placement.ClusterAffinity)
}

func (w PPWrapper) GetResourceSelectors() []v1alpha1.ResourceSelector {
	return ExtractResourceSelectors(w.Policy)
}

func (w CPPWrapper) GetResourceSelectors() []v1alpha1.ResourceSelector {
	return ExtractResourceSelectors(w.Policy)
}

func ExtractClusterNames(affinity *v1alpha1.ClusterAffinity) ([]string, bool) {
	if affinity == nil || len(affinity.ClusterNames) == 0 {
		return nil, false
	}
	return affinity.ClusterNames, true
}

func ExtractResourceSelectors(policy interface{}) []v1alpha1.ResourceSelector {
	switch p := policy.(type) {
	case *v1alpha1.PropagationPolicy:
		if p == nil || len(p.Spec.ResourceSelectors) == 0 {
			return nil
		}
		return p.Spec.ResourceSelectors
	case *v1alpha1.ClusterPropagationPolicy:
		if p == nil || len(p.Spec.ResourceSelectors) == 0 {
			return nil
		}
		return p.Spec.ResourceSelectors
	default:
		return nil
	}
}

func RemoveDuplicates(input []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, val := range input {
		if !seen[val] {
			seen[val] = true
			result = append(result, val)
		}
	}
	return result
}

func ExtractRelatedResources(verberClient client.ResourceVerber, resourceSelectors HasResourceSelectors) []string {
	relatedResources := make([]string, 0)
	for _, rs := range resourceSelectors.GetResourceSelectors() {
		getRes, getErr := verberClient.Get(strings.ToLower(rs.Kind), rs.Namespace, rs.Name)

		if errors.IsNotFound(getErr) {
			klog.Infof("not found resource %s/%s (%s). Skipping.", rs.Namespace, rs.Name, rs.Kind)
			continue
		}
		if getErr != nil {
			klog.Infof("failed to get resource %s/%s (%s): %v. Skipping.", rs.Namespace, rs.Name, rs.Kind, getErr)
			continue
		}
		if getRes == nil {
			klog.Infof("resource is nil without error for %s/%s (%s). Skipping.", rs.Namespace, rs.Name, rs.Kind)
			continue
		}

		relatedResources = append(relatedResources, fmt.Sprintf("%s/%s", rs.Namespace, rs.Name))
	}

	return relatedResources
}

func getGVKForKind(kind string) (*schema.GroupVersionKind, bool) {
	kindToGVK := map[string]schema.GroupVersionKind{
		"Deployment":  {Group: "apps", Version: "v1", Kind: "Deployment"},
		"StatefulSet": {Group: "apps", Version: "v1", Kind: "StatefulSet"},
		"DaemonSet":   {Group: "apps", Version: "v1", Kind: "DaemonSet"},
		"CronJob":     {Group: "batch", Version: "v1", Kind: "CronJob"},
		"Job":         {Group: "batch", Version: "v1", Kind: "Job"},
	}
	gvk, ok := kindToGVK[kind]
	if !ok {
		return nil, false
	}
	return &gvk, true
}

func formatAPIVersion(group, version string) string {
	if group == "" {
		return version
	}
	return fmt.Sprintf("%s/%s", group, version)
}

func BuildPolicyMetadata(meta *v1.Metadata) (*metav1.ObjectMeta, error) {
	labels, err := appcommon.ParseKeyValueStrings(meta.Labels)
	if err != nil {
		return nil, err
	}

	annotations, err := appcommon.ParseKeyValueStrings(meta.Annotations)
	if err != nil {
		return nil, err
	}

	return &metav1.ObjectMeta{
		Namespace:   meta.Namespace,
		Name:        meta.Name,
		Labels:      labels,
		Annotations: annotations,
	}, nil
}

func BuildResourceSelectors(reqs []v1.ResourceSelector, isClusterScoped bool) ([]v1alpha1.ResourceSelector, error) {
	var result []v1alpha1.ResourceSelector
	for _, req := range reqs {
		gvk, ok := getGVKForKind(req.Kind)
		if !ok {
			return nil, apperrors.UnsupportedResourceKind
		}

		/*  if req.Name == "" && len(req.LabelSelector) == 0 {
		      return nil, fmt.Errorf("either name or labelSelector must be set for kind %s", req.Kind)
		    }
		*/
		rs := v1alpha1.ResourceSelector{
			APIVersion: formatAPIVersion(gvk.Group, gvk.Version),
			Kind:       gvk.Kind,
			Name:       req.Name,
		}

		// ClusterPropagationPolicy는 namespace 지정 필요
		if isClusterScoped {
			rs.Namespace = req.Namespace
		}

		if len(req.LabelSelectors) > 0 {
			labels, err := appcommon.ParseKeyValueStrings(req.LabelSelectors)
			if err != nil {
				return nil, apperrors.InvalidKeyValueFormat
			}
			rs.LabelSelector = &metav1.LabelSelector{
				MatchLabels: labels,
			}
		}

		result = append(result, rs)
	}

	return result, nil
}

func BuildPlacement(p v1.Placement) (*v1alpha1.Placement, error) {
	if err := validatePlacement(p); err != nil {
		return nil, err
	}

	placement := &v1alpha1.Placement{
		ClusterAffinity: &v1alpha1.ClusterAffinity{
			ClusterNames: RemoveDuplicates(p.ClusterNames),
		},
	}

	rs := p.ReplicaScheduling
	if !isEmptyReplicaScheduling(rs) {
		placement.ReplicaScheduling = &v1alpha1.ReplicaSchedulingStrategy{
			ReplicaSchedulingType: rs.ReplicaSchedulingType,
		}
		// Only apply StaticWeightList if Divided
		if rs.ReplicaSchedulingType == v1alpha1.ReplicaSchedulingTypeDivided {
			placement.ReplicaScheduling.ReplicaDivisionPreference = rs.ReplicaDivisionPreference
			if rs.ReplicaDivisionPreference == v1alpha1.ReplicaDivisionPreferenceWeighted {
				var staticWeightList []v1alpha1.StaticClusterWeight
				for _, w := range rs.StaticWeightList {
					staticWeightList = append(staticWeightList, v1alpha1.StaticClusterWeight{
						TargetCluster: v1alpha1.ClusterAffinity{
							ClusterNames: RemoveDuplicates(w.TargetClusters),
						},
						Weight: w.Weight,
					})
				}
				if len(staticWeightList) > 0 {
					placement.ReplicaScheduling.WeightPreference = &v1alpha1.ClusterPreferences{
						StaticWeightList: staticWeightList,
					}
				}
			}
		}
	}

	return placement, nil
}

func validatePlacement(placement v1.Placement) error {
	rs := placement.ReplicaScheduling

	if len(placement.ClusterNames) < 1 {
		return apperrors.PolicyMissingTargetClusters
	}

	if isEmptyReplicaScheduling(rs) {
		return nil
	}

	if !isValidReplicaSchedulingType(rs.ReplicaSchedulingType) {
		return apperrors.UnsupportedPolicyReplicaSchedulingType
	}

	if rs.ReplicaSchedulingType != v1alpha1.ReplicaSchedulingTypeDivided {
		return nil
	}

	if !isValidReplicaDivisionPreference(rs.ReplicaDivisionPreference) {
		return apperrors.UnsupportedPolicyReplicaDivisionPreference
	}

	if rs.ReplicaDivisionPreference == v1alpha1.ReplicaDivisionPreferenceWeighted {
		return validateStaticWeightList(placement.ClusterNames, rs.StaticWeightList)
	}

	return nil
}

func isEmptyReplicaScheduling(rs *v1.ReplicaScheduling) bool {
	if rs == nil {
		return true
	}
	return rs.ReplicaSchedulingType == "" &&
		rs.ReplicaDivisionPreference == "" &&
		len(rs.StaticWeightList) == 0
}

func isValidReplicaSchedulingType(t v1alpha1.ReplicaSchedulingType) bool {
	return t == v1alpha1.ReplicaSchedulingTypeDuplicated ||
		t == v1alpha1.ReplicaSchedulingTypeDivided
}

func isValidReplicaDivisionPreference(p v1alpha1.ReplicaDivisionPreference) bool {
	return p == v1alpha1.ReplicaDivisionPreferenceAggregated ||
		p == v1alpha1.ReplicaDivisionPreferenceWeighted
}

func validateStaticWeightList(clusterNames []string, weights []v1.StaticWeight) error {
	affinitySet := make(map[string]struct{})
	for _, c := range clusterNames {
		affinitySet[c] = struct{}{}
	}

	for _, w := range weights {
		if len(w.TargetClusters) < 1 {
			return apperrors.EmptyStaticWeightClusters
		}

		for _, t := range w.TargetClusters {
			if _, ok := affinitySet[t]; !ok {
				return apperrors.InvalidStaticWeightClusters
			}
		}

		if w.Weight < 1 {
			return apperrors.InvalidPolicyReplicaWeight
		}
	}
	return nil
}

/*
func CompareRelatedClusters(policy HasClusterNames, managedClusters *domain.ClusterList) ClusterRelation {
	if userClusters.HasGlobalScope {
			return RelationEqual
		}

	ppClusters, exists := policy.GetClusterNames()
	if !exists {
		return RelationDisjoint
	}

	userSet := make(map[string]struct{})
	ppSet := make(map[string]struct{})
	matchCount := 0

	for _, c := range managedClusters.Items {
		userSet[c.FederatedClusterName] = struct{}{}
	}

	for _, c := range ppClusters {
		ppSet[c] = struct{}{}
	}

	for pp := range ppSet {
		if _, ok := userSet[pp]; ok {
			matchCount++
		}
	}

	if matchCount == 0 {
		return RelationDisjoint
	}

	if matchCount == len(ppSet) {
		return RelationEqual
	}

	return RelationPartial
}
*/
