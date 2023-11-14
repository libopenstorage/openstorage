# Changelog

## Releases

### v0.173.0 - (11/13/2023)

* Add DefragJob data structure

### v0.172.0 - (11/07/2023)

* Add new checksum flag to volume spec

### v0.171.0 - (11/02/2023)

* Add new CLI for filesystem check

### v0.170.0 - (10/26/2023)

* Add new CLI to trigger checksum validation

### v0.169.0 - (10/24/2023)

* Add quorum member flag to StorageNode

### v0.168.0 - (09/14/2023)

* Add new field cpu cores to api Node

### v0.167.0 - (09/14/2023)

* Add new field verbose to fastpath configuration

### v0.166.0 - (09/01/2023)

* Add pool rebalance schedule related APIs

### v0.165.0 - (08/30/2023)

* Add CloudDriveType field to StorageResource

### v0.164.0 - (07/17/2023)

* Add serverSideEncryption field to aws credentials

### v0.163.0 - (07/13/2023)

* Add mode field to SdkStorageRebalanceRequest

### v0.162.0 - (07/11/2023)

* Add new Watch endpoint

### v0.161.0 - (06/12/2023)

* Add auto-journal IO profile

### v0.160.0 - (06/12/2023)

* Add journal IO profile

### v0.159.0 - (04/06/2023)

* Add 3 new fields to the CloudBackupSize API:
  * total_download_bytes
  * compressed_object_bytes
  * capacity_required_for_restore

### v0.158.0 - (03/31/2023)

* Update stats for new VolumeBytesUsedByNode api

### v0.157.0 - (03/23/2023)

* Update stats for new VolumeBytesUsedByNode api

### v0.156.0 - (02/21/2023)

* Update NearSync clone state related fields

### v0.155.0 - (01/11/2023)

* Add NearSync related fields

### v0.154.0 - (12/21/2022)

* Adds additionalLabels field in SDK Clone request to pass additional labels to be added post-clone

### v0.153.0 - (11/09/2022)

* Add api for autofstrim push and pop

### v0.152.0 - (10/26/2022)

* FA DirectAccess CreateOptions is not honoroed by PX

### v0.151.0 - (9/13/2022)

* Add NONE value to sharedv4 service type enum

### v0.150.0 - (8/17/2022)

* PWX-26324: Extend fastpath config to disable auto fallback

### v0.149.0 - (8/4/2022)

* Add NFS Credential support

### v0.148.0 - (8/2/2022)

* Update SDK cloudBackup delete API

### v0.147.0 - (7/29/2022)

* Update volume spec for windows share

### v0.146.0 - (7/1/2022)

* Update create/delete bucket api param to include endpoint.

### v0.145.0 - (6/20/2022)

* Update create bucket api param to include anonymous bucket access mode.

### v0.144.0 - (6/8/2022)

* Revise for thin pool metadata device in storage resource

### v0.143.0 - (8/6/2022)

* Updating open storage bucket access credential object.

### v0.142.0 - (6/6/2022)

* Updating api params for Create and Delete Bucket service.

### v0.141.0 - (6/6/2022)

* Extend storage resource for thin pool metadata disk

### v0.140.0 - (5/31/2022)

* Added Access and Revoke api for SDK Bucket driver interface.

### v0.139.0 - (5/17/2022)

* Added Create and Delete api for SDK Bucket driver interface.

### v0.138.0 - (4/12/2022)

* Fixes many typos

### v0.137.0 - (3/31/2022)

* Adds full_vol_name field to PureBlockSpec.
* Adds full_vol_name field to PureFileSpec.

### v0.136.0 - (3/17/2022)

* Add topology requirement field the VolumeSpec.

### v0.135.0 - (2/7/2022)

* Add serial number to PureBlockSpec.

### v0.134.0 - (2/7/2022)

* Add scheduler topology field to the StorageNode object.

### v0.133.0 - (2/3/2022)

* Renumbered the fields to match release branches.

### v0.132.0 - (1/31/2022)

* Added a new cluster pairing mode for OneTimeMigration.

### v0.131.0 - (1/15/2022)

* Add a readahead flag in volume spec

### v0.130.0 - (1/4/2022)

* Add a filename field to DiagsCollectionRequest for test purposes

### v0.129.0 - (11/30/2021)

* Add message to show auto fstrim enable/disable info

### v0.128.0 - (11/17/2021)

* Add api for auto fstrim disk usage report

### v0.127.0 - (11/11/2021)

* Upgrade to proto3

### v0.126.0 - (09/28/2021)

* Added api for auto fs trim status

### v0.125.0 - (09/28/2021)

* Add sharedv4 failover strategy option

### v0.124.0 - (09/20/2021)

* Add pool maintenance status code

### v0.123.0 - (07/07/2021)

* Add DerivedIoProfile to Volume and a new NONE IO Profile

### v0.122.0 - (07/07/2021)

* Add a flag in Sharedv4ServiceSpec to indicate whether the service needs to be accessed outside of the cluster

### v0.121.0 - (06/09/2021)

* Remove unused fields from volume spec and VolumeState

### v0.120.0 - (05/12/2021)

* Add RelaxedReclaimPurge SDK API to OpenstorageNode service.

### v0.119.0 - (05/03/2021)

* Add new api to update existing credentials

### v0.118.0 - (04/22/2021)

* Add support for per-volume throttling of IOPS and/or bandwidth

### v0.117.0 - (05/06/2021)

* Change sharedv4 servicey type enum to conform to the style guilde

### v0.116.0 - (04/26/2021)

* Add Pure pass through volume specs into ProxySpec
* Parse mountOptions flag from storage class for Pure pass through volumes

### v0.115.0 - (04/12/2021)

* Add new fields to VolumeInfo and ReplicaSet to enable dynamic volume chunking

### v0.114.0 - (04/07/2021)

* StorageNode now have security status

### v0.113.0 - (04/06/2021)

* Add live option to diags collection SDK to collect live cores

### v0.112.0 - (03/30/2021)

* auto fstrim flag in volume spec

### v0.111.0 - (03/24/2021)

* Additions to CloudBackupEnumerate API to allow enumerating cloudbackups
* with whose source volumes are missing in the cluster. Also indicate if
* cloudbackup belongs to current cluster with enumerate data.

### v0.110.0 - (02/22/2021)

* Add SDK for diags collection

### v0.109.0 - (02/17/2021)

* Add Trashcan volume objects

### v0.108.0 - (01/26/2021)

* Handle volume spec update for fastpath

### v0.107.0 - (12/14/2020)

* Add a spec for defining a service for sharedv4 volumes. The service
  can be used for accessing this sharedv4 volume within and from outside
  the cluster.

### v0.106.0 - (01/06/2021)

* Extend volume stats structure to include discards

### v0.105.0 - (11/17/2020)

* Use destination instance ID for cloud driver transfer job

### v0.104.0 - (11/05/2020)

* Fastpath extend to carry node UUID instead of internal int

### v0.103.0 - (11/05/2020)

* Change the API definitions for OpenstorageJobServer RPCs

### v0.102.0 - (11/05/2020)

* Add CloudDriveTransfer job type

### v0.101.0 - (10/14/2020)

* Add CredentialDeleteReferences API

### v0.100.0 - (09/10/2020)

* Rename NodeDrain API to RemoveVolumeAttachments API
* Add OpenstorageJob service to query running and past jobs and their states.

### v0.99.0 - (9/10/2020)

* Add VolumeUsageByNode SDK API to OpenstorageNode service.

### v0.98.0 - (8/7/2020)

* Rename reflection volumes to proxy volumes.

### v0.97.0 - (8/7/2020)

* Add support for Reflection Volumes.
* Reflection Volumes essentially reflect an external data source as an openstorage volume.
* Added a new field ReflectionSpec to VolumeSpec object.

### v0.96.0 - (8/5/2020)

* Add CredentialId field to ClusterPairCreate api

### v0.95.0 - (8/4/2020)

* Removed IO_PROFILE_BKUPSRC from ioProfile list

### v0.94.0 - (7/29/2020)

* Added proxy write flag in volume spec

### v0.93.0 - (7/23/2020)

* Renamed FilesystemTrim Api GetStatus() to Status().

### v0.92.0 - (7/16/2020)

* Add sharedv4_mount_options field to Volume and VolumeSpec object.
* The sharedv4_mount_options will be used at runtime while mounting the sharedv4 volume from
  a node (client) which does not have the volume replica.

### v0.91.0 - (7/13/2020)

* Add mount_options field to Volume and VolumeSpec object.
* The mount_options will be used at runtime while mounting the volume.

### v0.90.0 - (7/8/2020)

* Added new field to CloudBackupGroupCreate api

### v0.89.0 - (7/7/2020)

* Remove LastUpdateTime from RebalanceJobSummary and added it RebalanceJob,

### v0.88.0 - (7/6/2020)

* Added new field to CloudBackupCreate api

### v0.87.0 - (6/29/2020)

* Modified fsck service interface and added new fields to volume and volume spec

### v0.86.0 - (6/25/2020)

* Add support for volume xattr update
*
### v0.85.0 - (6/25/2020)

* Updated rebalance data structures

### v0.84.0 - (6/24/2020)

* Updated rebalance data structures

### v0.83.0 - (6/16/2020)

* Added support for fetching cloud backup size

### v0.82.0 - (6/11/2020)

* Modified fsck service interface and added new fields to volume spec

### v0.81.0 - (6/3/2020)

* Add storage-class options to credentials

### v0.80.0 - (6/2/2020)

* Add direct_io as IO strategy

### v0.79.0 - (6/1/2020)

* Add "dirty" flag to fastpath volumes.

### v0.78.0 - (4/28/2020)

* Add SDK APIs for storage rebalance

### v0.77.0 - (4/22/2020)

* Add "deleteOnFailure" flag for snapshotGroup api

### v0.76.0 - (4/7/2020)

* Add implementation specific additional attributes for volume

### v0.75.0 - (3/18/2020)

* Add IAM flag for credentials

### v0.74.0 - (3/4/2020)

* Add VolumeCatalog api for volumes

### v0.73.0 - Tech Preview (1/27/2020)

* Add Restore volume spec for Cloud Backup restore api

### v0.72.0 - (1/21/2020)

* Added documentation to SdkRule about new denial support

### v0.71.0 - Tech Preview (1/15/2020)

* Add auto to IoProfile

### v0.70.0 - Tech Preview (1/9/2020)

* Add support for public user access

### v0.69.0 - Tech Preview (10/12/2019)

* Add support for filesystem trim background operation
* Add support for filesystem check background operation

### v0.68.0 - Tech Preview (11/01/2019)

* Add ExportSpec to Volume object

### v0.67.0 - Tech Preview (11/01/2019)

* Add proxy flag for credentials

### v0.66.0 - Tech Preview (10/22/2019)

* Add missing fields to cloudbackup schedule structure in sdk

### v0.65.0 - Tech Preview (09/10/2019)

* Add pool UUIDs in ReplicaSet object

### v0.64.0 - Tech Preview (9/26/2019)

* Deprecated StoragePool.ID. Please use StoragePool.Uuid

### v0.63.0 - Tech Preview (9/26/2019)

* Added new parameter to cloudbackup enumerate API to specify backup ID

### v0.62.0 - Tech Preview (9/26/2019)

* Added new api to resize storage pools
* Added new fields UUID and LastOperation to the StoragePool object

### v0.61.0 - Tech Preview (9/10/2019)

* Add fields last_attached and last_detached to the Volume object.

### v0.60.0 - Tech Preview (6/11/2019)

* Added new api for cloudbackup schedule update

### v0.59.0 - Tech Preview (7/16/2019)

* Add EnumerateWithFilters api for Node server which returns complete StorageNode object

### v0.58.0 - Tech Preview (6/5/2019)

* Add cloud group-backup API

### v0.57.0 - Tech Preview (6/4/2019)

* Added new param credential API to control path style access to s3

### v0.56.0 - Tech Preview (6/3/2019)

* Addition to Node details, to store the hardware type

### v0.55.0 - Tech Preview (5/7/2019)

* Additions to cloudbackup enumerate filters.

### v0.54.0 - Tech Preview (5/2/2019)

* Added new field FullBackupFrequency to cloudbackup create API

### v0.53.0 - Tech Preview (4/23/2019)

* Added new field RetentionDays to cloudbackup schedule

### v0.52.0 - Tech Preview (4/23/2019)

* (breaking change) Fix REST Gateway Snapshot enumerate with filters endpoint

### v0.51.0 - Tech Preview (4/11/2019)

* Added VolumeInspectOptions to OpenStorageVolume.Inspect
* Added new OpenStorageVolume.InspectWithFilters API

### v0.50.0 - Tech Preview (4/2/2019)

* Add groupId field to SdkCloudBackupStatus structure

### v0.49.0 - Tech Preview (4/3/2019)

* Add Group to VolumeLocator

### v0.48.0 - Tech Preview (4/2/2019)

* Add cluster pair Mode option in CreatePair and ProcessPair requests

### v0.47.0 - Tech Preview (3/26/2019)

* Change io_strategy type from oneof to pointer

### v0.46.0 - Tech Preview (3/26/2019)

* Handle spec update for nodiscard, io_strategy

### v0.45.0 - Tech Preview (3/13/2019)

* Add new APIs for managing OpenStorage ClusterDomains

### v0.44.0 - Tech Preview (3/21/2019)

* Add ownership support to OpenStorageStoragePolicy

### v0.43.0 - Tech Preview (3/12/2019)

* Add ownership support to OpenStorageCredential service APIs

### v0.42.0 - Tech Preview (2/20/2019)

* SnapEnumerate REST endpoint now accepts empty volume ids

### v0.41.0 - Tech Preview (2/20/2019)

* Add driver options to RPCs in the MountAttach service

### v0.40.0 - Tech Preview (2/19/2019)

* Storage policy support
* Allow Enforce/Release of storage policy

### v0.39.0 - Tech Preview (1/29/2019)

* Additional fields to cloud-backup data structure to track group cloud backups.

### v0.38.0 - Tech Preview (1/27/2019)

* Ownership reworked to gain access type control. Now it supports Read, Write,
  and Admin access types.

### v0.37.0 - Tech Preview (1/16/2019)

* Ownership support in the VolumeSpec

### v0.36.0 - Tech Preview (1/7/2019)

* Refactor confusing labels.
    * Deprecated Volume.Spec.VolumeLabels.
    * Any labels in Volume.Spec.VolumeLabels will be copied to Volume.Locator
    * Added Labels to Volume.Create
    * Volume.Update now takes Labels and Name instead of VolumeLocator
    * Volume.Inspect now also returns Name and Labels to match Volume.Create

### v0.35.0 - Tech Preview (1/4/2019)

* Rename SdkVolumeAttachRequest_Options to SdkVolumeAttachOptions
* Rename SdkVolumeUnmount_Options to SdkVolumeUnmountOptions
* Rename SdkVolumeDetach_Options to SdkVolumeDetachOptions
* Change SdkVolumeMountRequest to include SdkVolumeAttachOptions

### v0.34.0 - Tech Preview (1/2/2019)

* Role support
* Added Cluster Pair and Migrate to Capabilities since they were missing

### v0.33.0 - Tech Preview (12/05/2018)

* Add TaskId and ClusterId to CloudMigrate status request

### v0.32.0 - Tech Preview (11/28/2018)

* Removing unused objects created for cluster pair APIs

### v0.31.0 - Tech Preview (11/27/2018)

* (breaking change) REST API for Sdk OpenStorageAlerts has changed
* (breaking change) OpenStorageAlerts.Enumerate is now EnumerateWithFilters

### v0.30.0 - Tech Preview (11/20/2018)

* SDK Alerts enumerate chunking bug resolution.

### v0.29.0 - Tech Preview (11/17/2018)

* SDK Alerts enumerate is now a server side streaming api.

### v0.28.0 - Tech Preview (11/15/2018)

* (breaking change) Restructured all SDK REST routes
* (breaking change) Reworded OpenStorageCloudBackup.Enumerate to EnumerateWithFilters

### v0.27.0 - Tech Preview (11/8/2018)

* Add new API for extracting volume capacity usage details.

### v0.26.0 - Tech Preview (11/14/2018)

* Extend attribute of StorageResource to be marked as a cache.

### v0.25.0 - Tech Preview (11/13/2018)

* Added labels field to cloud backup create message

### v0.24.0 - Tech Preview (11/12/2018)

* Added ETA fields to cloud backup and cloud migrate status messages

### v0.23.0 - Tech Preview (11/2/2018)

* Cloud migrate status and cloud backup status now report
  total bytes to be transferred and bytes already transferred.
* These status blocks also report the start time of the operation
  so that client could calculate progress of the operation.

### v0.22.0 - Tech Preview (11/1/2018)

* Rename the field "name" to "TaskId" in sdkCloudBackupcreate/restore/status
  structures.

### v0.21.0 - Tech Preview (10/31/2018)

* Addition of ClusterPairing and VolumeMigrate services

### v0.20.0 - Tech Preview (11/1/2018)

* Added ETA for cloud snap status.

### v0.19.0 - Tech Preview (10/23/2018)

* CloudBackupStatus now returns CredentialUUID used for cloud for the
  backup/restore op under consideration.

### v0.18.0 - Tech Preview (10/23/2018)

* Following CloudBackup APIs have been refactored to include task id rather
  than source volume id.
* CloudBackupCreate now returns task id.
* CloudBackupRestore too returns task id along with restore volume id.
* CloudBackupStatusRequest can take task id as an optional parameter.
* Map key for CloudBackupStatusResponse is task id rather than source volume id.
* CloudBackupStateChange takes in taskid rather than source volume id.

### v0.17.0 - Tech Preview (10/21/2018)

* Added IoStrategy - ability to specify I/O characteristics.

### v0.16.0 - Tech Preview (10/15/2018)

* Changed value of SdkSchedulePolicyCreateRequest from `SchedulePolicy` to the
  correct name of `schedule_policy`. This will not impact Golang.

### v0.15.0 - Tech Preview (10/10/2018)

* Added support to set the snapshot schedule policy of a Volume

### v0.14.0 - Tech Preview (10/8/2018)

* Added support for periodic type in OpenStorageSchedulePolicy service

### v0.13.0 - Tech Preview (9/27/2018)

* Added new field to CloudBackup schedules that allows scheduled backups
  to be always full and never incremental.

### v0.12.0 - Tech Preview (9/27/2018)

* Moved MountAttach service REST endpoints to their own namespace
* Added new MountAttach to SdkSErviceCapability

### v0.11.0 - Tech Preview (9/25/2018)

* New RPC in service called OpenStorageAlerts has been created and documented to
  allow deleting alert events.

### v0.10.0 - Tech Preview (9/24/2018)

* New service called OpenStorageAlerts has been created and documented to
  allow querying alert events.

### v0.9.0 - Tech Preview (9/18/2018)

NOTE: This release has breaking chages for the Mount/Attach/Detach/Unmount calls

* New service called OpenStorageMountAttach has been created and documented to
  hold the mount/attach/detach/unmount calls.
* Mount/Attach/Detach/Unmount calls have been moved from the OpenStorageVolume
  service to the OpenStorageMountAttach service.

### v0.8.0 - Tech Preview (9/11/2018)

* SdkVolumeSnapshotEnumerateWithFilters all attributes are now optional. [#609](https://github.com/libopenstorage/openstorage/issues/609)

### v0.7.0 - Tech Preview (9/5/2018)

* Add `Name` to `StorageCluster`. This name will hold the name given to the cluster by the administrator. The `StorageCluster.Id` will now hold a unique id for the cluster.

### v0.6.0 - Tech Preview (8/30/2018)

* Remove unsupported FS Types from supported drivers [#593](https://github.com/libopenstorage/openstorage/issues/593)
* Remove SDK Alert calls as they will be redesinged [#596](https://github.com/libopenstorage/openstorage/issues/596)

### v0.5.0 - Tech Preview (8/25/2018)

* Added `queue_depth` to VolumeSpec and VolumeSpecUpdate
* Remove values from VolumeSpecUpdate which cannot be updated [#590](https://github.com/libopenstorage/openstorage/issues/590)

### v0.4.0 - Tech Preview (8/24/2018)

* Added bucket name and encryption key to SdkCredentialCreateRequest
* Added the ability to disable ssl connection to SdkAwsCredentialRequest

### v0.3.0 - Tech Preview (8/20/2018)

* Added SchedulerNodeName field to StorageNode object

### v0.2.0 - Tech Preview (8/16/2018)

* Changed Credentials.Create to take `name` as a required parameter

### Tech Preview (8/7/2018)

* Added [Identity.Version](https://libopenstorage.github.io/w/generated-api.html#methodopenstorageapiopenstorageidentityversion) Service

### Tech Preview (8/3/2018)

* Added [Idenity](https://libopenstorage.github.io/w/generated-api.html#serviceopenstorageapiopenstorageidentity) Service
* Added [Identity.Capabilities](https://libopenstorage.github.io/w/generated-api.html#methodopenstorageapiopenstorageidentitycapabilities) RPC
