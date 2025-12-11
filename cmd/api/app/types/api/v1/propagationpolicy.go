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

package v1

import "github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"

// PostPropagationPolicyRequest defines the request structure for creating a propagation policy.
// todo this is only a simple version of pp request, just for POC
type PostPropagationPolicyRequest struct {
	Metadata          Metadata           `json:"metadata" binding:"required"`
	ResourceSelectors []ResourceSelector `json:"resourceSelectors" binding:"required"`
	Placement         Placement          `json:"placement" binding:"required"`
}

type Metadata struct {
	Name                        string   `json:"name" binding:"required"`
	Namespace                   string   `json:"namespace"`
	Labels                      []string `json:"labels"`
	Annotations                 []string `json:"annotations"`
	PreserveResourcesOnDeletion *bool    `json:"preserveResourcesOnDeletion"`
}

type ResourceSelector struct {
	Kind           string   `json:"kind" binding:"required"`
	Namespace      string   `json:"namespace"`
	Name           string   `json:"name"`
	LabelSelectors []string `json:"labelSelectors"`
}

type Placement struct {
	ClusterNames      []string           `json:"clusterNames"`
	ReplicaScheduling *ReplicaScheduling `json:"replicaScheduling"`
}

type ReplicaScheduling struct {
	ReplicaSchedulingType     v1alpha1.ReplicaSchedulingType     `json:"replicaSchedulingType"`
	ReplicaDivisionPreference v1alpha1.ReplicaDivisionPreference `json:"replicaDivisionPreference"`
	StaticWeightList          []StaticWeight                     `json:"staticWeightList"`
}
type StaticWeight struct {
	TargetClusters []string `json:"targetClusters"`
	Weight         int64    `json:"weight"`
}

// PostPropagationPolicyResponse defines the response structure for creating a propagation policy.
type PostPropagationPolicyResponse struct {
}

// PutPropagationPolicyRequest defines the request structure for updating a propagation policy.
type PutPropagationPolicyRequest struct {
	PropagationData string `json:"propagationData" binding:"required"`
	IsClusterScope  bool   `json:"isClusterScope"`
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
}

// PutPropagationPolicyResponse defines the response structure for updating a propagation policy.
type PutPropagationPolicyResponse struct {
}

// DeletePropagationPolicyRequest defines the request structure for deleting a propagation policy.
type DeletePropagationPolicyRequest struct {
	IsClusterScope bool   `json:"isClusterScope"`
	Namespace      string `json:"namespace"`
	Name           string `json:"name" binding:"required"`
}

// DeletePropagationPolicyResponse defines the response structure for deleting a propagation policy.
type DeletePropagationPolicyResponse struct {
}
