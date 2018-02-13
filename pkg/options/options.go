package options

import (
	"strconv"
)

// Options specifies keys from a key-value pair
// that can be passed in to the APIS
const (
	// OptionsSecret Secret used to get secret for secure devices
	OptionsSecret = "SECRET_NAME"
	// OptionsSecretKey Key used to secure devices
	OptionsSecretKey = "SECRET_KEY"
	// OptionsSecretContext Context used to provide additional info to retrive secret
	OptionsSecretContext = "SECRET_CONTEXT"
	// OptionsUnmountBeforeDetach Issue an Unmount before trying the detach
	OptionsUnmountBeforeDetach = "UNMOUNT_BEFORE_DETACH"
	// OptionsDeleteAfterUnmount Delete the mount path after Unmount
	OptionsDeleteAfterUnmount = "DELETE_AFTER_UNMOUNT"
	// OptionsDeleteAfterUnmount Introduce a delay before deleting mount path
	OptionsWaitBeforeDelete = "WAIT_BEFORE_DELETE"
	// OptionsRedirectDetach Redirect detach to the node where volume is attached
	OptionsRedirectDetach = "REDIRECT_DETACH"
	// OptionsDeviceFuseMount name of fuse mount device
	OptionsDeviceFuseMount = "DEV_FUSE_MOUNT"
	// OptionsForceDetach Forcefully detach device from kernel
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
