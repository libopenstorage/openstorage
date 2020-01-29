# Changelog

## Releases
### v0.69.1 - Tech Preview (1/9/2020)

* Add support for public user access

### v0.69.0 - Tech Preview (10/12/2019)

* Add support for filesystem trim background operation
* Add support for filesystem check background operation

### v0.68.0 - Tech Preview (11/01/2019)

* Add ExportSpec to Volume object

### v0.67.0 - Tech Preview (10/29/2019)

* Add proxy flag for credentials

### v0.66.0 - Tech Preview (10/22/2019)

* Add missing fields to cloudbackup schedule structure in sdk

### v0.65.0 - Tech Preview (09/10/2019)

* Add pool UUIDs in ReplicaSet object

### v0.64.0 - Tech Preview (9/26/2019)

* Deprecated StoragePool.ID. Please use StoragePool.Uuid

### v0.63.0 - Tech Preview (9/26/2019)

* Deprecated StoragePool.ID. Please use StoragePool.Uuid

### v0.62.0 - Tech Preview (9/26/2019)

* Added new parameter to cloudbackup enumerate API to specify backup ID

### v0.61.0 - Tech Preview (9/26/2019)

* Added new api to resize storage pools
* Added new fields UUID and LastOperation to the StoragePool object

### v0.60.0 - Tech Preview (6/11/2019)

* Added new api for cloudbackup schedule update

### v0.59.0 - Tech Preview (7/16/2019)

* Add EnumerateWithFilters api for Node server which returns complete StorageNode object

### v0.58.0 - Tech Preview (6/5/2019)

* Add cloud group-backup API

### v0.57.0 - Tech Preview (6/4/2019)

* Addition to Node details, to store the hardware type

### v0.56.0 - Tech Preview (6/4/2019)

* Added new param credential API to control path style access to s3

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
* Added Cluster Pair and Migrate to Capabilites since they were missing

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
