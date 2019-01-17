package v1alpha1

import (
	crdv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

// RuleActionType is a type for actions that are supported in a stork rule
type RuleActionType string

const (
	// RuleActionCommand is a command action
	RuleActionCommand RuleActionType = "command"
	// ClusterPairResourceName is name for "clusterpair" resource
	ClusterPairResourceName = "clusterpair"
	// ClusterPairResourcePlural is plural for "clusterpair" resource
	ClusterPairResourcePlural = "clusterpairs"
	// MigrationResourceName is name for "migration" resource
	MigrationResourceName = "migration"
	// MigrationResourcePlural is plural for "migration" resource
	MigrationResourcePlural = "migrations"
	// GroupSnapshotResourceName is name for "groupvolumesnapshot" resource
	GroupSnapshotResourceName = "groupvolumesnapshot"
	// GroupSnapshotResourcePlural is plural for the "groupvolumesnapshot" resource
	GroupSnapshotResourcePlural = "groupvolumesnapshots"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Rule denotes an object to declare a rule that performs actions on pods
type Rule struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            []RuleItem `json:"spec"`
}

// RuleItem represents one items in a stork rule spec
type RuleItem struct {
	// PodSelector is a map of key value pairs that are used to select the pods using their labels
	PodSelector map[string]string `json:"podSelector"`
	// Actions are actions to be performed on the pods selected using the selector
	Actions []RuleAction `json:"actions"`
}

// RuleAction represents an action in a stork rule item
type RuleAction struct {
	// Type is a type of the stork rule action
	Type RuleActionType `json:"type"`
	// Background indicates that the action needs to be performed in the background
	// +optional
	Background bool `json:"background,omitempty"`
	// RunInSinglePod indicates that the action needs to be performed in a single pod
	//                from the list of pods that match the selector
	// +optional
	RunInSinglePod bool `json:"runInSinglePod,omitempty"`
	// Value is the actual action value for e.g the command to run
	Value string `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleList is a list of stork rules
type RuleList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []Rule `json:"items"`
}

// ClusterPairSpec is the spec to create the cluster pair
type ClusterPairSpec struct {
	Config  api.Config        `json:"config"`
	Options map[string]string `json:"options"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPair represents pairing with other clusters
type ClusterPair struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            ClusterPairSpec   `json:"spec"`
	Status          ClusterPairStatus `json:"status,omitempty"`
}

// ClusterPairStatusType is the status of the pair
type ClusterPairStatusType string

const (
	// ClusterPairStatusInitial is the initial state when pairing is created
	ClusterPairStatusInitial ClusterPairStatusType = ""
	// ClusterPairStatusPending for when pairing is still pending
	ClusterPairStatusPending ClusterPairStatusType = "Pending"
	// ClusterPairStatusReady for when pair is ready
	ClusterPairStatusReady ClusterPairStatusType = "Ready"
	// ClusterPairStatusError for when pairing is in error state
	ClusterPairStatusError ClusterPairStatusType = "Error"
	// ClusterPairStatusDegraded for when pairing is degraded
	ClusterPairStatusDegraded ClusterPairStatusType = "Degraded"
	// ClusterPairStatusDeleting for when pairing is being deleted
	ClusterPairStatusDeleting ClusterPairStatusType = "Deleting"
)

// ClusterPairStatus is the status of the cluster pair
type ClusterPairStatus struct {
	// Status of the pairing with the scheduler
	// +optional
	SchedulerStatus ClusterPairStatusType `json:"schedulerStatus"`
	// Status of pairing with the storage driver
	// +optional
	StorageStatus ClusterPairStatusType `json:"storageStatus"`
	// ID of the remote storage which is paired
	// +optional
	RemoteStorageID string `json:"remoteStorageId"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPairList is a list of cluster pairs
type ClusterPairList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ClusterPair `json:"items"`
}

// MigrationSpec is the spec used to migrate apps between clusterpairs
type MigrationSpec struct {
	ClusterPair       string            `json:"clusterPair"`
	Namespaces        []string          `json:"namespaces"`
	IncludeResources  bool              `json:"includeResources"`
	StartApplications bool              `json:"startApplications"`
	Selectors         map[string]string `json:"selectors"`
}

// MigrationStatus is the status of a migration operation
type MigrationStatus struct {
	Stage     MigrationStageType  `json:"stage"`
	Status    MigrationStatusType `json:"status"`
	Resources []*ResourceInfo     `json:"resources"`
	Volumes   []*VolumeInfo       `json:"volumes"`
}

// ResourceInfo is the info for the migration of a resource
type ResourceInfo struct {
	Name                  string `json:"name"`
	Namespace             string `json:"namespace"`
	meta.GroupVersionKind `json:",inline"`
	Status                MigrationStatusType `json:"status"`
	Reason                string              `json:"reason"`
}

// VolumeInfo is the info for the migration of a volume
type VolumeInfo struct {
	PersistentVolumeClaim string              `json:"persistentVolumeClaim"`
	Namespace             string              `json:"namespace"`
	Volume                string              `json:"volume"`
	Status                MigrationStatusType `json:"status"`
	Reason                string              `json:"reason"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Migration represents migration status
type Migration struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            MigrationSpec   `json:"spec"`
	Status          MigrationStatus `json:"status"`
}

// MigrationStatusType is the status of the migration
type MigrationStatusType string

const (
	// MigrationStatusInitial is the initial state when migration is created
	MigrationStatusInitial MigrationStatusType = ""
	// MigrationStatusPending for when migration is still pending
	MigrationStatusPending MigrationStatusType = "Pending"
	// MigrationStatusCaptured for when migration specs have been captured
	MigrationStatusCaptured MigrationStatusType = "Captured"
	// MigrationStatusInProgress for when migration is in progress
	MigrationStatusInProgress MigrationStatusType = "InProgress"
	// MigrationStatusFailed for when migration has failed
	MigrationStatusFailed MigrationStatusType = "Failed"
	// MigrationStatusPartialSuccess for when migration was partially successful
	MigrationStatusPartialSuccess MigrationStatusType = "PartialSuccess"
	// MigrationStatusSuccessful for when migration has completed successfully
	MigrationStatusSuccessful MigrationStatusType = "Successful"
)

// MigrationStageType is the stage of the migration
type MigrationStageType string

const (
	// MigrationStageInitial for when migration is created
	MigrationStageInitial MigrationStageType = ""
	// MigrationStageVolumes for when volumes are being migrated
	MigrationStageVolumes MigrationStageType = "Volumes"
	// MigrationStageApplications for when applications are being migrated
	MigrationStageApplications MigrationStageType = "Applications"
	// MigrationStageFinal is the final stage for migration
	MigrationStageFinal MigrationStageType = "Final"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MigrationList is a list of Migrations
type MigrationList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []Migration `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GroupVolumeSnapshot represents a group snapshot
type GroupVolumeSnapshot struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            GroupVolumeSnapshotSpec   `json:"spec"`
	Status          GroupVolumeSnapshotStatus `json:"status"`
}

// GroupVolumeSnapshotSpec represents the spec for a group snapshot
type GroupVolumeSnapshotSpec struct {
	// PreSnapshotRule is the name of rule applied before taking the snapshot. The rule needs to be
	// in the same namespace as the group volumesnapshot
	PreSnapshotRule string `json:"preSnapshotRule"`
	// PreSnapshotRule is the name of rule applied after taking the snapshot. The rule needs to be
	// in the same namespace as the group volumesnapshot
	PostSnapshotRule string `json:"postSnapshotRule"`
	// PVCSelector selects the PVCs that are part of the group snapshot
	PVCSelector PVCSelectorSpec `json:"pvcSelector"`
	// Options are pass-through parameters that are passed to the driver handling the group snapshot
	Options map[string]string `json:"options"`
}

// PVCSelectorSpec is the spec to select the PVCs for group snapshot
type PVCSelectorSpec struct {
	meta.LabelSelector
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GroupVolumeSnapshotList is a list of group volume snapshots
type GroupVolumeSnapshotList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []GroupVolumeSnapshot `json:"items"`
}

// GroupVolumeSnapshotStatus is status for the group snapshot
type GroupVolumeSnapshotStatus struct {
	Stage           GroupVolumeSnapshotStageType  `json:"stage"`
	Status          GroupVolumeSnapshotStatusType `json:"status"`
	VolumeSnapshots []*VolumeSnapshotStatus       `json:"volumeSnapshots"`
}

// VolumeSnapshotStatus captures the status of a volume snapshot operation
type VolumeSnapshotStatus struct {
	VolumeSnapshotName string
	TaskID             string
	ParentVolumeID     string
	DataSource         *crdv1.VolumeSnapshotDataSource
	Conditions         []crdv1.VolumeSnapshotCondition
}

// GroupVolumeSnapshotStatusType is types of statuses of a group snapshot operation
type GroupVolumeSnapshotStatusType string

const (
	// GroupSnapshotInitial is when the group snapshot is created and no action has yet been performed
	GroupSnapshotInitial GroupVolumeSnapshotStatusType = ""
	// GroupSnapshotPending is when the group snapshot is in pending state waiting for another event
	GroupSnapshotPending GroupVolumeSnapshotStatusType = "Pending"
	// GroupSnapshotInProgress is when the group snapshot is in progress
	GroupSnapshotInProgress GroupVolumeSnapshotStatusType = "InProgress"
	// GroupSnapshotFailed is when the group snapshot has failed
	GroupSnapshotFailed GroupVolumeSnapshotStatusType = "Failed"
	// GroupSnapshotSuccessful is when the group snapshot has succeeded
	GroupSnapshotSuccessful GroupVolumeSnapshotStatusType = "Successful"
)

// GroupVolumeSnapshotStageType is the stage of the group snapshot
type GroupVolumeSnapshotStageType string

const (
	// GroupSnapshotStageInitial is when the group snapshot is just created
	GroupSnapshotStageInitial GroupVolumeSnapshotStageType = ""
	// GroupSnapshotStagePreChecks is when the group snapshot is going through prechecks
	GroupSnapshotStagePreChecks GroupVolumeSnapshotStageType = "PreChecks"
	// GroupSnapshotStagePreSnapshot is when the pre-snapshot rule is executing for the group snapshot
	GroupSnapshotStagePreSnapshot GroupVolumeSnapshotStageType = "PreSnapshot"
	// GroupSnapshotStageSnapshot is when the snapshots are being taken for the group snapshot
	GroupSnapshotStageSnapshot GroupVolumeSnapshotStageType = "Snapshot"
	// GroupSnapshotStagePostSnapshot is when the post-snapshot rule is executing for the group snapshot
	GroupSnapshotStagePostSnapshot GroupVolumeSnapshotStageType = "PostSnapshot"
	// GroupSnapshotStageFinal is when all stages are done for the group snapshot
	GroupSnapshotStageFinal GroupVolumeSnapshotStageType = "Final"
)
