package metrics

import (
	"encoding/json"
	"fmt"
	v1api "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/nats-io/nats.go"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type ClusterUsage struct {
	Time                v1.Time  `json:"time"`
	HostClusterStatus   Status   `json:"hostClusterStatus"`
	MemberClusterStatus []Status `json:"memberClusterStatus"`
}

type Status struct {
	ClusterId     string             `json:"clusterId"`
	Name          string             `json:"name"`
	UID           k8stypes.UID       `json:"uid"`
	Status        string             `json:"status"`
	NodeSummary   *v1api.NodeSummary `json:"nodeSummary"`
	RealTimeUsage v1api.Usage        `json:"realTimeUsage"`
	RequestUsage  v1api.Usage        `json:"requestUsage"`
}

// GetClustersRealTimeUsage retrieves the latest cluster metric snapshot from JetStream KV.
func GetClustersRealTimeUsage() (*ClusterUsage, error) {
	raw, err := fetchClusterMetricsFromKV()
	if err != nil {
		klog.Errorf("Failed to fetch cluster metrics: %v", err)
		return nil, err
	}

	var usage *ClusterUsage
	if err := json.Unmarshal(raw, &usage); err != nil {
		klog.Errorf("Failed to unmarshal cluster metrics: %v", err)
		return nil, err
	}

	return usage, nil
}

// fetchClusterMetricsFromKV connects to JetStream and retrieves the cluster.metrics key from the KV bucket.
func fetchClusterMetricsFromKV() ([]byte, error) {
	nc, err := connectToNATS()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := nc.Drain(); err != nil {
			klog.Warningf("NATS drain error: %v", err)
		}
	}()

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	kv, err := js.KeyValue(Env.NatsBucketName)
	if err == nats.ErrBucketNotFound {
		klog.Infof("KV bucket %q is not found", Env.NatsBucketName)
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access KV bucket %q: %w", Env.NatsBucketName, err)
	}

	entry, err := kv.Get(Env.NatsSubjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %q: %w", Env.NatsSubjectName, err)
	}

	klog.Infof("Retrieved cluster metric value: %s", string(entry.Value()))
	return entry.Value(), nil
}

// connectToNATS establishes a connection to the NATS server using basic authentication.
func connectToNATS() (*nats.Conn, error) {
	opts := []nats.Option{
		nats.UserInfo(Env.NatsUsername, Env.NatsPassword),
	}

	nc, err := nats.Connect(Env.NatsUrl, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	return nc, nil
}
