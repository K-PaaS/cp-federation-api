package v1

import (
	"github.com/karmada-io/dashboard/pkg/common/types"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

type ResourceYaml struct {
	Namespace string       `json:"namespace,omitempty"`
	Name      string       `json:"name"`
	UID       k8stypes.UID `json:"uid"`
	Yaml      string       `json:"yaml"`
}

type ManifestRequest struct {
	Data string `json:"data" binding:"required"`
}

type ResourceList struct {
	ListMeta  types.ListMeta `json:"listMeta"`
	Resources []Resource     `json:"resources"`
}

type Resource struct {
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	Policy    PolicyMeta        `json:"policy"`
}

type PolicyMeta struct {
	IsClusterScope bool   `json:"isClusterScope"`
	Name           string `json:"name"`
}
