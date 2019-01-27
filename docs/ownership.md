# Ownership

Ownership is the method used by the SDK to manage ACL access to a resource.

## Volume Ownership

Ownership is set by the following:

* User creates a new volume
* User creates a clone from a volume which they do not own, but have access to
* Ownership information is carried along snapshots

### Volume API access types
The following rules apply to access levels.

* Volume owners have admin level automatically
* Only owners or admin can edit acls

The following access types are required to access the shown Volume service
Apis:

* Admin:
    * Delete
* Write:
    * SnapshotRestore, SnapshotScheduleUpdate
    * Update, Attach, Detach, Mount, Unmount
* Read
    * The rest of the volumd SDK API calls not above
    * Cloud backup calls use this access type also

