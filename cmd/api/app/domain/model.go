package domain

type ClusterCredential struct {
	ClusterID    string
	APIServerURL string
	BearerToken  string
}

type ClusterList struct {
	Items []Cluster `json:"items"`
}
type Cluster struct {
	ClusterID            string `json:"clusterId"`
	Name                 string `json:"name"`
	FederatedClusterUID  string `json:"federatedClusterUID"`
	FederatedClusterName string `json:"federatedClusterName"`
	IsFederated          bool   `json:"isFederated"`
}

type FederationRequest struct {
	ClusterID            string `json:"clusterId"`
	FederatedClusterUID  string `json:"federatedClusterUID,omitempty"`
	FederatedClusterName string `json:"federatedClusterName,omitempty"`
}
