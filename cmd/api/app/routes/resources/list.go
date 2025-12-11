package resources

import (
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/pkg/client"
	"github.com/karmada-io/dashboard/pkg/common/types"
	"github.com/karmada-io/dashboard/pkg/dataselect"
	"github.com/karmada-io/dashboard/pkg/resource/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

var (
	ppKindMeta  = policyKindMeta{"propagationpolicy", "propagationpolicy.karmada.io/permanent-id", false}
	cppKindMeta = policyKindMeta{"clusterpropagationpolicy", "clusterpropagationpolicy.karmada.io/permanent-id", true}
)

type policyKindMeta struct {
	Kind           string
	LabelKey       string
	IsClusterScope bool
}

// ------------------------ List ------------------------

func GetResourceList(verber client.ResourceVerber, kind string, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*v1.ResourceList, error) {
	klog.Infof("Getting list of %s", kind)
	list, err := verber.List(kind, nsQuery.ToRequestParam())
	if err != nil {
		klog.ErrorS(err, "Failed to get resource list", "namespace", nsQuery.ToRequestParam(), "kind", kind)
		return nil, apperrors.ResourceError(err)
	}
	return toResourceList(verber, list.Items, dsQuery), nil
}

func toResourceList(verber client.ResourceVerber, unstructured []unstructured.Unstructured, dsQuery *dataselect.DataSelectQuery) *v1.ResourceList {
	resourceList := &v1.ResourceList{
		Resources: make([]v1.Resource, 0),
		ListMeta:  types.ListMeta{TotalItems: len(unstructured)},
	}
	resourceCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(unstructured), dsQuery)
	unstructured = fromCells(resourceCells)
	resourceList.ListMeta = types.ListMeta{TotalItems: filteredTotal}

	for _, u := range unstructured {
		resourceList.Resources = append(resourceList.Resources, toResource(&u))
	}

	err := AttachPolicyMetaToResources(verber, resourceList.Resources)
	if err != nil {
		klog.ErrorS(err, "failed to attach policyMeta to resources")
	}
	return resourceList
}

func toResource(u *unstructured.Unstructured) v1.Resource {
	labels := make(map[string]string)
	if u.GetLabels() != nil {
		labels = u.GetLabels()
	}
	return v1.Resource{
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
		Labels:    labels,
	}
}

func AttachPolicyMetaToResources(verber client.ResourceVerber, resourceList []v1.Resource) error {
	ppMap, err := BuildPolicyMetaMapByPermanentId(verber, ppKindMeta)
	if err != nil {
		return err
	}

	cppMap, err := BuildPolicyMetaMapByPermanentId(verber, cppKindMeta)
	if err != nil {
		return err
	}

	allMap := mergePolicyMetaMaps(ppMap, cppMap)

	for i := range resourceList {
		labels := resourceList[i].Labels
		if len(labels) == 0 {
			continue
		}

		if meta, ok := getPolicyMeta(labels, ppKindMeta.LabelKey, allMap); ok {
			resourceList[i].Policy = meta
		} else if meta, ok := getPolicyMeta(labels, cppKindMeta.LabelKey, allMap); ok {
			resourceList[i].Policy = meta
		}

	}

	return nil
}

func BuildPolicyMetaMapByPermanentId(verber client.ResourceVerber, meta policyKindMeta) (map[string]v1.PolicyMeta, error) {
	policies, err := verber.List(meta.Kind, "")
	if err != nil {
		klog.ErrorS(err, "Failed to get policies", "kind", meta.Kind)
		return nil, err
	}

	policyMetaMap := make(map[string]v1.PolicyMeta)
	for _, p := range policies.Items {
		labels := p.GetLabels()
		if len(labels) == 0 {
			continue
		}

		if pid, ok := labels[meta.LabelKey]; ok {
			policyMetaMap[pid] = v1.PolicyMeta{
				IsClusterScope: meta.IsClusterScope,
				Name:           p.GetName(),
			}
		}
	}

	return policyMetaMap, nil
}

func mergePolicyMetaMaps(maps ...map[string]v1.PolicyMeta) map[string]v1.PolicyMeta {
	result := make(map[string]v1.PolicyMeta)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func getPolicyMeta(labels map[string]string, key string, policyMap map[string]v1.PolicyMeta) (v1.PolicyMeta, bool) {
	if pid, ok := labels[key]; ok {
		if v, exist := policyMap[pid]; exist {
			return v, true
		}
	}
	return v1.PolicyMeta{}, false
}
