package cronjob

import (
	"context"
	utils "github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/common/errors"
	"github.com/karmada-io/dashboard/pkg/common/helpers"
	"github.com/karmada-io/dashboard/pkg/resource/common"
	v1 "k8s.io/api/batch/v1"
	client "k8s.io/client-go/kubernetes"
	"log"
)

func GetCustomCronJobList(client client.Interface, nsQuery *common.NamespaceQuery) (*v1.CronJobList, error) {
	log.Println("Getting list of cronjobs")
	cronjobs, err := client.BatchV1().CronJobs(nsQuery.ToRequestParam()).List(context.TODO(), helpers.ListEverything)
	_, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return cronjobs, nil
}

func GetCronJobNames(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	cronjobs, err := GetCustomCronJobList(client, nsQuery)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, d := range cronjobs.Items {
		names = append(names, d.Name)
	}
	return names, nil
}
func GetCronJobLabelSelectors(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	cronjobs, err := GetCustomCronJobList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.CronJob
	for _, c := range cronjobs.Items {
		objs = append(objs, &c)
	}
	labels := utils.ExtractLabels(objs)

	return labels, nil
}

func GetNamespacesWithCronJob(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	cronjobs, err := GetCustomCronJobList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.CronJob
	for _, c := range cronjobs.Items {
		objs = append(objs, &c)
	}
	namespaces := utils.ExtractNamespaces(objs)

	return namespaces, nil
}
