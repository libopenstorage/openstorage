package vsphere

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/vsphere/vclib"
)

// derived from https://github.com/kubernetes/kubernetes/blob/release-1.9/pkg/cloudprovider/providers/vsphere/vsphere_util.go#L94
// VSphereConfig represents the vsphere configuration
type VSphereConfig struct {
	// User is the vCenter username.
	User string
	// Password is the vCenter password in clear text.
	Password string
	// VCenterIP is the vcenter IP to connect on
	VCenterIP string
	// VCenterPort is the vcenter port to connect on
	VCenterPort string
	// InsecureFlag True if vCenter uses self-signed cert.
	InsecureFlag bool
	// RoundTripperCount is the Soap round tripper count (retries = RoundTripper - 1)
	RoundTripperCount uint
	// VMUUID is the VM Instance UUID of virtual machine which can be retrieved from instanceUuid
	// property in VmConfigInfo, or also set as vc.uuid in VMX file.
	// If not set, will be fetched from the machine via sysfs (requires root)
	VMUUID string
}

// Represents a vSphere instance where one or more kubernetes nodes are running.
type VSphereInstance struct {
	conn *vclib.VSphereConnection
}

var datastoreFolderIDMap = make(map[string]map[string]string)

// Structure that represents Virtual Center configuration
type VirtualCenterConfig struct {
	// vCenter username.
	User string
	// vCenter password in clear text.
	Password string
	// vCenter port
	VCenterPort string
	// Datacenter in which VMs are located.
	Datacenter string
	// Soap round tripper count (retries = RoundTripper - 1)
	RoundTripperCount uint
}

func ReadVSphereConfigFromEnv() (*VSphereConfig, error) {
	var cfg VSphereConfig
	var err error

	cfg.VCenterIP, err = storageops.GetEnvValueStrict("VSPHERE_VCENTER")
	if err != nil {
		return nil, err
	}
	cfg.VCenterPort, err = storageops.GetEnvValueStrict("VSPHERE_VCENTER_PORT")
	if err != nil {
		return nil, err
	}
	cfg.User, err = storageops.GetEnvValueStrict("VSPHERE_USER")
	if err != nil {
		return nil, err
	}
	cfg.Password, err = storageops.GetEnvValueStrict("VSPHERE_PASSWORD")
	if err != nil {
		return nil, err
	}

	cfg.InsecureFlag = false
	insecure, err := storageops.GetEnvValueStrict("VSPHERE_INSECURE")
	if err == nil && strings.ToLower(insecure) == "true" {
		cfg.InsecureFlag = true
	}

	cfg.VMUUID, _ = storageops.GetEnvValueStrict("VSPHERE_VM_UUID")

	return &cfg, nil
}

// IsDevMode checks if requirement env variables are set to run the pkg outside vsphere in dev mode
func IsDevMode() bool {
	_, err := storageops.GetEnvValueStrict("VSPHERE_VM_UUID")
	if err != nil {
		return false
	}

	_, err = storageops.GetEnvValueStrict("VSPHERE_TEST_DATASTORE")
	return err == nil
}

// Get canonical volume path for volume Path.
// Borrowed from https://github.com/kubernetes/kubernetes/blob/release-1.10/pkg/cloudprovider/providers/vsphere/vsphere_util.go#L312
// Example1: The canonical path for volume path - [vsanDatastore] kubevols/volume.vmdk will be [vsanDatastore] 25d8b159-948c-4b73-e499-02001ad1b044/volume.vmdk
// Example2: The canonical path for volume path - [vsanDatastore] 25d8b159-948c-4b73-e499-02001ad1b044/volume.vmdk will be same as volume Path.
func getCanonicalVolumePath(ctx context.Context, dc *vclib.Datacenter, volumePath string) (string, error) {
	var folderID string
	var folderExists bool
	canonicalVolumePath := volumePath
	dsPathObj, err := vclib.GetDatastorePathObjFromVMDiskPath(volumePath)
	if err != nil {
		return "", err
	}
	dsPath := strings.Split(strings.TrimSpace(dsPathObj.Path), "/")
	if len(dsPath) <= 1 {
		return canonicalVolumePath, nil
	}
	datastore := dsPathObj.Datastore
	dsFolder := dsPath[0]
	folderNameIDMap, datastoreExists := datastoreFolderIDMap[datastore]
	if datastoreExists {
		folderID, folderExists = folderNameIDMap[dsFolder]
	}

	// Get the datastore folder ID if datastore or folder doesn't exist in datastoreFolderIDMap
	if !datastoreExists || !folderExists {
		if !vclib.IsValidUUID(dsFolder) {
			dummyDiskVolPath := "[" + datastore + "] " + dsFolder + "/" + dummyDiskName
			// Querying a non-existent dummy disk on the datastore folder.
			// It would fail and return an folder ID in the error message.
			_, err := dc.GetVirtualDiskPage83Data(ctx, dummyDiskVolPath)
			if err != nil {
				re := regexp.MustCompile("File (.*?) was not found")
				match := re.FindStringSubmatch(err.Error())
				canonicalVolumePath = match[1]
			}
		}
		diskPath := vclib.GetPathFromVMDiskPath(canonicalVolumePath)
		if diskPath == "" {
			return "", fmt.Errorf("Failed to parse canonicalVolumePath: %s in getcanonicalVolumePath method", canonicalVolumePath)
		}
		folderID = strings.Split(strings.TrimSpace(diskPath), "/")[0]
		setdatastoreFolderIDMap(datastoreFolderIDMap, datastore, dsFolder, folderID)
	}

	canonicalVolumePath = strings.Replace(volumePath, dsFolder, folderID, 1)
	if filepath.Base(datastore) != datastore {
		// If datastore is within cluster, add cluster path to the volumePath
		canonicalVolumePath = strings.Replace(canonicalVolumePath, filepath.Base(datastore), datastore, 1)
	}
	return canonicalVolumePath, nil
}

func setdatastoreFolderIDMap(
	datastoreFolderIDMap map[string]map[string]string,
	datastore string,
	folderName string,
	folderID string) {
	folderNameIDMap := datastoreFolderIDMap[datastore]
	if folderNameIDMap == nil {
		folderNameIDMap = make(map[string]string)
		datastoreFolderIDMap[datastore] = folderNameIDMap
	}
	folderNameIDMap[folderName] = folderID
}

// GetStoragePodMoList fetches the managed storage pod objects for the given references
//		Only the properties is the given property list will be populated in the response
func GetStoragePodMoList(
	ctx context.Context,
	client *vim25.Client,
	storagePodRefs []types.ManagedObjectReference,
	properties []string) ([]mo.StoragePod, error) {
	var storagePodMoList []mo.StoragePod
	pc := property.DefaultCollector(client)
	err := pc.Retrieve(ctx, storagePodRefs, properties, &storagePodMoList)
	if err != nil {
		logrus.Errorf("Failed to get Storagepod managed objects from storage pod refs: %+v, properties: %+v, err: %v",
			storagePodRefs, properties, err)
		return nil, err
	}
	return storagePodMoList, nil
}
