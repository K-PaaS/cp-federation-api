package daemonset

import (
	"context"
	utils "github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/dashboard/pkg/common/helpers"
	"github.com/karmada-io/dashboard/pkg/resource/common"
	v1 "k8s.io/api/apps/v1"
	client "k8s.io/client-go/kubernetes"
	"log"
)

func GetCustomDaemonSetList(client client.Interface, nsQuery *common.NamespaceQuery) (*v1.DaemonSetList, error) {
	log.Println("Getting list of daemonsets")
	ds, err := client.AppsV1().DaemonSets(nsQuery.ToRequestParam()).List(context.TODO(), helpers.ListEverything)
	_, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return ds, nil
}

func GetDaemonSetNames(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	ds, err := GetCustomDaemonSetList(client, nsQuery)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, d := range ds.Items {
		names = append(names, d.Name)
	}
	return names, nil
}
func GetDaemonSetLabelSelectors(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	ds, err := GetCustomDaemonSetList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.DaemonSet
	for _, d := range ds.Items {
		objs = append(objs, &d)
	}
	labels := utils.ExtractLabels(objs)

	return labels, nil
}

func GetNamespacesWithDaemonSet(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	ds, err := GetCustomDaemonSetList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.DaemonSet
	for _, d := range ds.Items {
		objs = append(objs, &d)
	}
	namespaces := utils.ExtractNamespaces(objs)

	return namespaces, nil
}
