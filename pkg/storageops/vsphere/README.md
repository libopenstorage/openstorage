### To test vSphere

You will first need to have a vSphere environment and then provide details of the vcenter server and VM as below.

```bash
export VSPHERE_VCENTER=my-vc.demo.com
export VSPHERE_VCENTER_PORT=443
export VSPHERE_USER=administrator@vsphere.local
export VSPHERE_PASSWORD=my-vc-password
export VSPHERE_INSECURE=true

# To get VSPHERE_VM_UUID, go to https://<vcenter-ip>/mob/?moid=<VM-MOREF>&doPath=config and get the "uuid" field value
# To get the VM-MOREF, select the VM in vcenter server and you will see a string of format "VirtualMachine:vm-155" in the URL. vm-155 is the moref.

export VSPHERE_VM_UUID=42124a20-d049-9c0a-0094-1552b320fb18
export VSPHERE_TEST_DATASTORE=<test-datastore-to-use>

# VSPHERE_TEST_DATASTORE above can be a vSphere datastore or datastore cluster name. When testing changes, it is recommended to test with both.

go test -v
```

