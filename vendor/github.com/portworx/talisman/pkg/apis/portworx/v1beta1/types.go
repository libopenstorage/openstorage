package v1beta1

import (
	"github.com/libopenstorage/openstorage/api"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnforcementType Defines the types of enforcement on the given rules
type EnforcementType string

const (
	// EnforcementRequired specifies that the rule is required and must be strictly enforced
	EnforcementRequired EnforcementType = "required"
	// EnforcementPreferred specifies that the rule is preferred and can be best effort
	EnforcementPreferred EnforcementType = "preferred"
)

// AffinityRuleType specifies the type an affinity rule can take
type AffinityRuleType string

const (
	// Affinity means the rule specifies an affinity to objects that match the below label selector requirements
	Affinity AffinityRuleType = "affinity"
	// AntiAffinity means the rule specifies an anti-affinity to objects that match the below label selector requirements
	AntiAffinity AffinityRuleType = "antiAffinity"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster describes a Portworx cluster
type Cluster struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            ClusterSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList is a list of Cluster objects in Kubernetes
type ClusterList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []Cluster `json:"items"`
}

// ClusterSpec defines the specification for a Cluster
type ClusterSpec struct {
	// Kvdb is the key value store configuration
	Kvdb KvdbSpec `json:"kvdb"`
	// PXImage is the Portworx image to use on all nodes of the cluster.
	// +optional
	PXImage string `json:"pxImage,omitempty"`
	// PXTag is the Portworx docker image tag
	// +optional
	PXTag string `json:"pxTag"`
	// OCIMonImage is the docker image for OCI monitor that runs on each k8s node
	// +optional
	OCIMonImage string `json:"ociMonImage"`
	// OCIMonTag is the docker tag for OCI monitor
	// +optional
	OCIMonTag string `json:"ociMonTag"`
	// Network specifies the networking setting to be used for all nodes. This
	// can be overridden by individual nodes in the NodeSpec
	Network NodeNetwork `json:"network,omitempty"`
	// Storage specifies the storage configuration to be used for all nodes.
	// This can be overridden by individual nodes in the NodeSpec
	Storage StorageSpec `json:"storage,omitempty"`
	// Placement specifies the rules by which PX nodes are selected
	Placement PlacementSpec `json:"placement,omitempty"`
	// Env is the list of environment variables to expose to PX pods
	Env []v1.EnvVar `json:"env,omitempty"`
}

// Nodes are all Portworx nodes participating in this cluster

// KvdbSpec defines the kvdb configuration
type KvdbSpec struct {
	// Endpoints is the list of kvdb endpoints
	Endpoints []string `json:"endpoints"`
	// BasicAuthSecret is the secret contain username and password for basic auth
	BasicAuthSecret string `json:"accessSecret,omitempty"`
	// CertificateSecret is the secret that contains the cert files required for etcd auth
	CertificateSecret string `json:"certificateSecret,omitempty"`
	// ACLTokenSecret is the secret name containing the ACL token for consul auth
	ACLTokenSecret string `json:"aclTokenSecret,omitempty"`
}

// ClusterStatus is the status of the Portworx cluster
type ClusterStatus struct {
	StatusInfo
	Name         string       `json:"name,omitempty"`
	NodeStatuses []NodeStatus `json:"nodeStatuses,omitempty"`
}

// NodeStatus represents status of a cluster node
type NodeStatus struct {
	StatusInfo
	Name string `json:"name,omitempty"`
}

// StatusInfo is used to represent the status of any entity in the cluster
type StatusInfo struct {
	Ready bool       `json:"ready"`
	Code  api.Status `json:"code"`
	// The following follow the same definition as PodStatus
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// NodeNetwork specifies which network interfaces the Node should use for data
// and management transport
type NodeNetwork struct {
	Data string `json:"data"`
	Mgmt string `json:"mgmt"`
}

// StorageSpec specifies the storage configuration for a node
type StorageSpec struct {
	Devices             []string `json:"devices,omitempty"`
	ZeroStorage         bool     `json:"zeroStorage,omitempty"`
	Force               bool     `json:"force,omitempty"`
	UseAll              bool     `json:"useAll,omitempty"`
	UseAllWithParitions bool     `json:"useAllWithParitions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PlacementSpec defines placement rules for various px components
type PlacementSpec struct {
	meta.TypeMeta `json:",inline"`
	PX            Placement `json:"px,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Placement encapsulates the various kubernetes options that control where pods are scheduled and executed.
type Placement struct {
	meta.TypeMeta   `json:",inline"`
	NodeAffinity    *v1.NodeAffinity    `json:"nodeAffinity,omitempty"`
	PodAffinity     *v1.PodAffinity     `json:"podAffinity,omitempty"`
	PodAntiAffinity *v1.PodAntiAffinity `json:"podAntiAffinity,omitempty"`
	Tolerations     []v1.Toleration     `json:"tolerations,omitemtpy"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumePlacementStrategy specifies a spec for volume placement in the cluster
type VolumePlacementStrategy struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            VolumePlacementSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumePlacementStrategyList is a list of VolumePlacementStrategy objects
type VolumePlacementStrategyList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	// Items are the list of volume placements strategy items
	Items []VolumePlacementStrategy `json:"items"`
}

// VolumePlacementSpec specifies a set of rules for volume placement in the cluster
type VolumePlacementSpec struct {
	// Rules defines a list of rules as part of the placement spec. All the rules specified will
	// be applied for volume placement.
	// Rules that have enforcement as "required" are strictly enforced while "preferred" are best effort.
	// In situations, where 2 or more rules conflict, the weight of the rules will dictate which wins.
	Rules []VolumePlacementRule `json:"rules"`
}

// VolumePlacementRule defines the rule for placing volume replicas
type VolumePlacementRule struct {
	// AffectedReplicas defines the number of volume replicas affected by this rule. If not provided,
	// rule would affect all replicas
	// (optional)
	AffectedReplicas int64 `json:"affectedReplicas,omitempty"`
	// Weight defines the weight of the rule which allows to break the tie with other matching rules. A rule with
	// higher weight wins over a rule with lower weight.
	// (optional)
	Weight int64 `json:"weight,omitempty"`
	// Enforcement specifies the rule enforcement policy. Can take values: required or preferred.
	// (optional)
	Enforcement EnforcementType `json:"enforcement,omitempty"`
	// Type is the type of the affinity rule
	Type AffinityRuleType `json:"type,omitempty"`
	// MatchExpressions is a list of label selector requirements. The requirements are ANDed.
	MatchExpressions []*meta.LabelSelectorRequirement `json:"matchExpressions,omitempty"`
}
