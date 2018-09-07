package options

import (
	"strconv"
)

// Options specifies keys from a key-value pair
// that can be passed in to the APIS
const (
	// OptionsSecret is an option provided to the following Openstorage Volume API
	// - Attach
	// It indicates the name of the secret stored in a secret store
	// SECRET_NAME in case of hashicorp's Vault will be the key from the key-value pair stored in its kv backend.
	// SECRET_NAME for Kubernetes secret, it is the name of the secret object itself
	OptionsSecret = "SECRET_NAME"
	// OptionsSecretKey is an option provided to the following Openstorage Volume API
	// - Attach
	// SECRET_KEY in case of kubernetes will be the key stored in the kubernetes secret with name SECRET_NAME
	OptionsSecretKey = "SECRET_KEY"
	// OptionsSecretContext is an option provided to the following Openstorage Volume API
	// - Attach
	// It indicates the additional context which could be used to retrieve the secret
	// SECRET_CONTEXT in case of kubernetes secret is the namespace in which the secret is created
	OptionsSecretContext = "SECRET_CONTEXT"
	// OptionsUnmountBeforeDetach is an option provided to the following Openstorage Volume API
	// - Detach
	// It indicates the Volume Driver to issue an Unmount before trying the detach operation
	OptionsUnmountBeforeDetach = "UNMOUNT_BEFORE_DETACH"
	// OptionsDeleteAfterUnmount is an option provided to the following Openstorage Volume API
	// - Unmount
	// It indicates the Volume Driver to delete the mount path after a successful Unmount
	OptionsDeleteAfterUnmount = "DELETE_AFTER_UNMOUNT"
	// OptionsWaitBeforeDelete is an option provided to the following Openstorage Volume API
	// - Unmount
	// This option is used in conjunction with OptionsDeleteAfterUnmount.
	// It indicates the Volume Driver to introduce a delay before deleting mount path
	OptionsWaitBeforeDelete = "WAIT_BEFORE_DELETE"
	// OptionsRedirectDetach is an option provided to the following Openstorage Volume API
	// - Detach
	// It indicates the Volume Driver to redirect detach to the node where volume is attached
	OptionsRedirectDetach = "REDIRECT_DETACH"
	// OptionsDeviceFuseMount is an option provided to the following Openstorage Volume APIs
	// - Mount
	// - Unmount
	// It is used for volume types which use FUSE mounts.
	// It provides the Volume Driver with the underlying name of fuse mount device
	OptionsDeviceFuseMount = "DEV_FUSE_MOUNT"
	// OptionsForceDetach is an option provided to the following Openstorage Volume API
	// - Detach
	// It indicates the Volume Driver to forcefully detach device from kernel
	OptionsForceDetach = "FORCE_DETACH"
)

func IsBoolOptionSet(options map[string]string, key string) bool {
	if options != nil {
		if value, ok := options[key]; ok {
			if b, err := strconv.ParseBool(value); err == nil {
				return b
			}
		}
	}

	return false
}
