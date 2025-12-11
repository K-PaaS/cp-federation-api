package job

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

func GetCustomJobList(client client.Interface, nsQuery *common.NamespaceQuery) (*v1.JobList, error) {
	log.Println("Getting list of jobs")
	jobs, err := client.BatchV1().Jobs(nsQuery.ToRequestParam()).List(context.TODO(), helpers.ListEverything)
	_, criticalError := errors.ExtractErrors(err)
	if criticalError != nil {
		return nil, criticalError
	}
	return jobs, nil
}

func GetJobNames(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	jobs, err := GetCustomJobList(client, nsQuery)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for _, d := range jobs.Items {
		names = append(names, d.Name)
	}
	return names, nil
}
func GetJobLabelSelectors(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	jobs, err := GetCustomJobList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.Job
	for _, j := range jobs.Items {
		objs = append(objs, &j)
	}

	labels := utils.ExtractLabels(objs)
	return labels, nil
}

func GetNamespacesWithJob(client client.Interface, nsQuery *common.NamespaceQuery) ([]string, error) {
	jobs, err := GetCustomJobList(client, nsQuery)
	if err != nil {
		return nil, err
	}

	var objs []*v1.Job
	for _, j := range jobs.Items {
		objs = append(objs, &j)
	}

	namespaces := utils.ExtractNamespaces(objs)
	return namespaces, nil
}
