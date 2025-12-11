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

package cluster

import (
	"context"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/metrics"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	karmadascheme "github.com/karmada-io/karmada/pkg/generated/clientset/versioned/scheme"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"log"

	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/dashboard/pkg/common/helpers"
	"github.com/karmada-io/dashboard/pkg/common/types"
	"github.com/karmada-io/dashboard/pkg/dataselect"
	"github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

const (
	AnnotationKeyClusterId = "cp.internal.federation.api/cluster-id"
)

// CustomCluster the definition of a cluster.
type CustomCluster struct {
	ClusterId         string                   `json:"clusterId"`
	Name              string                   `json:"name"`
	UID               k8stypes.UID             `json:"uid,omitempty"`
	KubernetesVersion string                   `json:"kubernetesVersion,omitempty"`
	Status            string                   `json:"status"`
	NodeSummary       *v1.NodeSummary          `json:"nodeSummary"`
	SyncMode          v1alpha1.ClusterSyncMode `json:"syncMode"`
	RealTimeUsage     v1.Usage                 `json:"realTimeUsage"`
	RequestUsage      v1.Usage                 `json:"requestUsage"`
}

type CustomClusterYaml struct {
	ClusterId string       `json:"clusterId"`
	Name      string       `json:"name"`
	UID       k8stypes.UID `json:"uid,omitempty"`
	Yaml      string       `json:"yaml"`
}

// CustomClusterList contains a list of clusters.
type CustomClusterList struct {
	ListMeta types.ListMeta  `json:"listMeta"`
	Clusters []CustomCluster `json:"clusters"`
	// List of non-critical errors, that occurred during resource retrieval.
	//	Errors []error `json:"errors"`
}

// GetCustomClusterList returns a list of all Nodes in the cluster.
func GetCustomClusterList(client karmadaclientset.Interface, dsQuery *dataselect.DataSelectQuery, managedClustersMap map[string]string, metricsOpt *metrics.ClusterUsage) (*CustomClusterList, error) {
	clusters, err := client.ClusterV1alpha1().Clusters().List(context.TODO(), helpers.ListEverything)
	nonCriticalErrors, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}

	// (add)
	// filter to managed cluster
	// add clusterId to annotations
	var filteredClusters []v1alpha1.Cluster
	for _, c := range clusters.Items {
		clusterID, exists := managedClustersMap[string(c.UID)]
		if !exists {
			continue
		}
		if c.Annotations == nil {
			c.Annotations = map[string]string{}
		}
		c.Annotations[AnnotationKeyClusterId] = clusterID
		filteredClusters = append(filteredClusters, c)
	}

	return toCustomClusterList(client, filteredClusters, nonCriticalErrors, dsQuery, metricsOpt), nil
}

func toCustomClusterList(_ karmadaclientset.Interface, clusters []v1alpha1.Cluster, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery, metricsOpt *metrics.ClusterUsage) *CustomClusterList {
	clusterList := &CustomClusterList{
		Clusters: make([]CustomCluster, 0),
		ListMeta: types.ListMeta{TotalItems: len(clusters)},
		//Errors:   nonCriticalErrors,
	}
	clusterCells, filteredTotal := dataselect.GenericDataSelectWithFilter(
		toCells(clusters),
		dsQuery,
	)
	clusters = fromCells(clusterCells)
	clusterList.ListMeta = types.ListMeta{TotalItems: filteredTotal}
	for _, cluster := range clusters {
		clusterList.Clusters = append(clusterList.Clusters, toCustomCluster(&cluster))
	}

	setClusterRealTimeUsage(clusterList.Clusters, metricsOpt)
	return clusterList
}

func toCustomCluster(cluster *v1alpha1.Cluster) CustomCluster {
	allocatedResources, err := getclusterAllocatedResources(cluster)
	if err != nil {
		log.Printf("Couldn't get allocated resources of %s cluster: %s\n", cluster.Name, err)
	}

	clusterId := ""
	if cluster.Annotations != nil {
		clusterId = cluster.Annotations[AnnotationKeyClusterId]
	}

	// nodeSummary
	nodeSummary := &v1.NodeSummary{
		ReadyNum: cluster.Status.NodeSummary.ReadyNum,
		TotalNum: cluster.Status.NodeSummary.TotalNum,
	}

	//requestUsage
	requestUsage := v1.Usage{
		CPU:    common.RoundToTwoDecimals(allocatedResources.CPUFraction),
		Memory: common.RoundToTwoDecimals(allocatedResources.MemoryFraction),
	}

	return CustomCluster{
		ClusterId:         clusterId,
		Name:              types.NewObjectMeta(cluster.ObjectMeta).Name,
		UID:               types.NewObjectMeta(cluster.ObjectMeta).UID,
		KubernetesVersion: cluster.Status.KubernetesVersion,
		Status:            getCustomClusterConditionStatus(cluster),
		NodeSummary:       nodeSummary,
		SyncMode:          cluster.Spec.SyncMode,
		RequestUsage:      requestUsage,
		RealTimeUsage:     v1.InitUsage(),
	}
}

// GetCustomClusterDetail gets details of cluster.
func GetCustomClusterDetail(client karmadaclientset.Interface, clusterName string) (*v1alpha1.Cluster, error) {
	log.Printf("Getting details of %s cluster", clusterName)
	cluster, err := client.ClusterV1alpha1().Clusters().Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func getCustomClusterConditionStatus(cluster *v1alpha1.Cluster) string {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == "Ready" {
			return GetConditionStatus(string(condition.Status))
		}
	}
	return "unknown"
}

func GetConditionStatus(conditionStatus string) string {
	switch conditionStatus {
	case string(metav1.ConditionTrue):
		return "ready"
	case string(metav1.ConditionFalse):
		return "not ready"
	case string(metav1.ConditionUnknown):
		return "unknown"
	default:
		return "unknown"
	}
}

func GetCustomClusterYaml(client karmadaclientset.Interface, clusterId string, clusterName string) (*CustomClusterYaml, error) {
	log.Printf("Getting yaml of %s cluster", clusterName)
	cluster, err := GetCustomClusterDetail(client, clusterName)
	if err != nil {
		klog.ErrorS(err, "GetCustomClusterDetail failed")
		switch {
		case apierrors.IsNotFound(err):
			return nil, apperrors.ClusterNotFoundInKarmada
		default:
			return nil, err
		}
	}

	yaml, err := common.EncodeToYAML(cluster, karmadascheme.Scheme)
	if err != nil {
		return nil, err
	}

	return &CustomClusterYaml{
		ClusterId: clusterId,
		Name:      cluster.Name,
		UID:       cluster.UID,
		Yaml:      yaml,
	}, nil
}
