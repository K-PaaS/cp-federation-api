package statefulset

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

func GetCustomStatefulSetList(client client.Interface, nsQuery *common.NamespaceQuery) (*v1.StatefulSetList, error) {
	log.Println("Getting list of statefulsets")
	sts, err := client.AppsV1().StatefulSets(nsQuery.ToRequestParam()).List(context.TODO(), helpers.ListEverything)
	_, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return sts, nil
}

func GetStatefulSetNames(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	sts, err := GetCustomStatefulSetList(client, nsQuery)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, d := range sts.Items {
		names = append(names, d.Name)
	}
	return names, nil
}
func GetStatefulSetLabelSelectors(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	sts, err := GetCustomStatefulSetList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.StatefulSet
	for _, s := range sts.Items {
		objs = append(objs, &s)
	}
	labels := utils.ExtractLabels(objs)

	return labels, nil
}

func GetNamespacesWithStatefulSet(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	sts, err := GetCustomStatefulSetList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.StatefulSet
	for _, s := range sts.Items {
		objs = append(objs, &s)
	}
	namespaces := utils.ExtractNamespaces(objs)

	return namespaces, nil
}
