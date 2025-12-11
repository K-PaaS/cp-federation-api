package deployment

import (
	"context"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	appv1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	utils "github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/dashboard/pkg/common/helpers"
	"github.com/karmada-io/dashboard/pkg/common/types"
	"github.com/karmada-io/dashboard/pkg/dataselect"
	"github.com/karmada-io/dashboard/pkg/resource/common"
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

type CustomDeploymentList struct {
	ListMeta    types.ListMeta     `json:"listMeta"`
	Deployments []CustomDeployment `json:"deployments"`
}

type CustomDeployment struct {
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	policies  string            `json:"policy"`
}

func GetAppsV1DeploymentList(client client.Interface, nsQuery *common.NamespaceQuery) (*v1.DeploymentList, error) {
	deployments, err := client.AppsV1().Deployments(nsQuery.ToRequestParam()).List(context.TODO(), helpers.ListEverything)
	_, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return deployments, nil
}

func GetAppsV1Deployment(client client.Interface, namespace string, deploymentName string) (*v1.Deployment, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metaV1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get deployment")
		return nil, err
	}
	return deployment, nil
}

func GetDeploymentNames(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	deployments, err := GetAppsV1DeploymentList(client, nsQuery)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, d := range deployments.Items {
		/*
			remove := false
			for _, prefix := range intra.Env.FilterNamespaces {
				if strings.HasPrefix(d.Namespace, prefix) {
					remove = true
					break
				}
			}
			if remove == false {
		*/
		names = append(names, d.Name)
		/*
			}
		*/
	}
	return names, nil
}
func GetDeploymentLabelSelectors(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	deployments, err := GetAppsV1DeploymentList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.Deployment
	for _, d := range deployments.Items {
		objs = append(objs, &d)
	}

	labels := utils.ExtractLabels(objs)
	return labels, nil
}

func GetNamespacesWithDeployment(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	deployments, err := GetAppsV1DeploymentList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.Deployment
	for _, d := range deployments.Items {
		objs = append(objs, &d)
	}

	namespaces := utils.ExtractNamespaces(objs)
	return namespaces, nil
}

func GetCustomDeploymentList(client client.Interface, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*CustomDeploymentList, error) {
	deployment, err := GetAppsV1DeploymentList(client, nsQuery)
	if err != nil {
		return nil, err
	}
	return toCustomDeploymentList(deployment.Items, dsQuery), nil
}

func toCustomDeploymentList(deployments []v1.Deployment, dsQuery *dataselect.DataSelectQuery) *CustomDeploymentList {
	deploymentList := &CustomDeploymentList{
		Deployments: make([]CustomDeployment, 0),
		ListMeta:    types.ListMeta{TotalItems: len(deployments)},
	}

	deploymentCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(deployments), dsQuery)
	deployments = fromCells(deploymentCells)
	deploymentList.ListMeta = types.ListMeta{TotalItems: filteredTotal}

	for _, deployment := range deployments {
		deploymentList.Deployments = append(deploymentList.Deployments, toCustomDeployment(&deployment))
	}

	return deploymentList
}

func toCustomDeployment(deployment *v1.Deployment) CustomDeployment {
	return CustomDeployment{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
		Labels:    deployment.Labels,
	}
}

func GetDeploymentYaml(client client.Interface, namespace string, deploymentName string) (*appv1.ResourceYaml, error) {
	deployment, err := GetAppsV1Deployment(client, namespace, deploymentName)
	if err != nil {
		return nil, apperrors.ResourceError(err)
	}

	yaml, err := utils.EncodeToYAML(deployment, k8sscheme.Scheme)
	if err != nil {
		return nil, err
	}

	return &appv1.ResourceYaml{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
		UID:       deployment.UID,
		Yaml:      yaml,
	}, nil

}
