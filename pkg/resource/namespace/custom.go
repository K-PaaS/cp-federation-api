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
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	appv1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	utils "github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/dashboard/pkg/common/helpers"
	"github.com/karmada-io/dashboard/pkg/common/types"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

type CustomNamespaceList struct {
	ListMeta   types.ListMeta    `json:"listMeta"`
	Namespaces []CustomNamespace `json:"namespaces"`
}

type CustomNamespace struct {
	Name                string            `json:"name"`
	Labels              map[string]string `json:"labels"`
	Status              v1.NamespacePhase `json:"status"`
	Created             string            `json:"created"`
	SkipAutoPropagation bool              `json:"skipAutoPropagation"`
}

// GetCustomNamespaceList returns a list of all namespaces in the cluster.
func GetCustomNamespaceList(client client.Interface) (*v1.NamespaceList, error) {
	klog.Infof("Getting list of namespaces")
	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), helpers.ListEverything)
	_, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return namespaces, nil
}

// GetCustomNamespace gets namespace.
func GetCustomNamespace(client client.Interface, name string) (*v1.Namespace, error) {
	klog.Infof("Getting of %s namespace", name)
	namespace, err := client.CoreV1().Namespaces().Get(context.TODO(), name, metaV1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get namespace")
		return nil, err
	}
	return namespace, nil
}

func GetNamespaceNames(client client.Interface) ([]string, error) {
	namespace, err := GetCustomNamespaceList(client)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, d := range namespace.Items {
		names = append(names, d.Name)
	}
	return names, nil
}
func GetNamespaceLabelSelectors(client client.Interface) ([]string, error) {
	namespace, err := GetCustomNamespaceList(client)
	if err != nil {
		return nil, err
	}

	var objs []*v1.Namespace
	for _, n := range namespace.Items {
		objs = append(objs, &n)
	}

	labels := utils.ExtractLabels(objs)
	return labels, nil
}

func DeleteNamespace(ctx context.Context, client client.Interface, name string) error {
	klog.Infof("Deleting Namespace %s", name)
	err := client.CoreV1().Namespaces().Delete(ctx, name, metaV1.DeleteOptions{})
	if err != nil {
		return apperrors.ResourceError(err)
	}

	return nil
}

func GetNamespaceYaml(client client.Interface, name string) (*appv1.ResourceYaml, error) {
	namespace, err := GetCustomNamespace(client, name)
	if err != nil {
		return nil, apperrors.ResourceError(err)
	}

	yaml, err := utils.EncodeToYAML(namespace, k8sscheme.Scheme)
	if err != nil {
		return nil, err
	}

	return &appv1.ResourceYaml{
		Name: namespace.Name,
		UID:  namespace.UID,
		Yaml: yaml,
	}, nil

}

func toCustomNamespace(namespace v1.Namespace) CustomNamespace {
	_, exist := namespace.Labels[skipAutoPropagationLabel]
	return CustomNamespace{
		Name:                namespace.Name,
		Labels:              namespace.Labels,
		Status:              namespace.Status.Phase,
		Created:             namespace.CreationTimestamp.Format("2006-01-02 15:04:05"),
		SkipAutoPropagation: exist,
	}
}
