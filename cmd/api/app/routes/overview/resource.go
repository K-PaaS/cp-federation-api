package overview

import (
	"context"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/pkg/client"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// GetCustomClusterResourceStatus returns the status of cluster resources.
func GetCustomClusterResourceStatus() *v1.ClusterResourceStatus {
	clusterResourceStatus := &v1.ClusterResourceStatus{}
	ctx := context.TODO()
	karmadaClient := client.InClusterKarmadaClient()
	kubeClient := client.InClusterClientForKarmadaAPIServer()

	clusterResourceStatus.PropagationPolicyNum = getPropagationPolicyNum(ctx, karmadaClient)
	clusterResourceStatus.NamespaceNum = getNamespaceNum(ctx, kubeClient)
	clusterResourceStatus.WorkloadNum = getWorkloadNum(ctx, kubeClient)
	clusterResourceStatus.ServiceNum = getServiceNum(ctx, kubeClient)
	clusterResourceStatus.ConfigNum = getConfigMapNum(ctx, kubeClient)
	return clusterResourceStatus
}

func getPropagationPolicyNum(ctx context.Context, karmadaClient karmadaclientset.Interface) int {
	var propagationPolicyNum int

	clusterPPRet, err := karmadaClient.PolicyV1alpha1().ClusterPropagationPolicies().List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("clusterPropagationPolicies", err)
	}
	propagationPolicyNum += len(clusterPPRet.Items)

	ppRet, err := karmadaClient.PolicyV1alpha1().PropagationPolicies("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("propagationPolicies", err)
	}
	propagationPolicyNum += len(ppRet.Items)

	return propagationPolicyNum
}

func getNamespaceNum(ctx context.Context, kubeClient kubernetes.Interface) int {
	nsRet, err := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("namespaces", err)
	}
	return len(nsRet.Items)
}

func getWorkloadNum(ctx context.Context, kubeClient kubernetes.Interface) int {
	// (deployment, statefulset, daemonset, cronjob, job)
	var workloadNum int
	deploymentRet, err := kubeClient.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("deployments", err)
	}
	workloadNum += len(deploymentRet.Items)

	statefulSetRet, err := kubeClient.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("statefulSets", err)
	}
	workloadNum += len(statefulSetRet.Items)

	daemonSetRet, err := kubeClient.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("daemonSets", err)
	}
	workloadNum += len(daemonSetRet.Items)

	cronJobRet, err := kubeClient.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("cronJobs", err)
	}
	workloadNum += len(cronJobRet.Items)

	jobRet, err := kubeClient.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("jobs", err)
	}
	workloadNum += len(jobRet.Items)

	return workloadNum
}

func getServiceNum(ctx context.Context, kubeClient kubernetes.Interface) int {
	var serviceNum int
	svcRet, err := kubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("services", err)
	}
	serviceNum += len(svcRet.Items)

	ingressRet, err := kubeClient.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("ingresses", err)
	}
	serviceNum += len(ingressRet.Items)

	return serviceNum
}

func getConfigMapNum(ctx context.Context, kubeClient kubernetes.Interface) int {
	var configMapNum int

	secretRet, err := kubeClient.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("secrets", err)
	}
	configMapNum += len(secretRet.Items)
	cmRet, err := kubeClient.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return failedNum("configmaps", err)
	}
	configMapNum += len(cmRet.Items)

	return configMapNum
}

func failedNum(resource string, err error) int {
	klog.Warningf("get %s failed. err: %v", resource, err)
	return -1
}
