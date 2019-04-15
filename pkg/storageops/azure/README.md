### To test Azure

You will first need to create a Azure instance and then provide details of this instance as below.

If you are running the tests on an instance under scale set, only then you need to provide `AZURE_SCALE_SET_NAME`. Also you will have to provide the instance id (index in scale set) for `AZURE_INSTANCE_ID` instead of the instance name.

```bash
export AZURE_INSTANCE_ID=<instance-id>
export AZURE_SCALE_SET_NAME=<scale-set-name>
export AZURE_SUBSCRIPTION_ID=<subscription-id>
export AZURE_RESOURCE_GROUP_NAME=<resource-group-name-of-instance>
export AZURE_ENVIRONMENT=<azure-cloud-environment>
export AZURE_TENANT_ID=<tenant-id>
export AZURE_CLIENT_ID=<client-id>
export AZURE_CLIENT_SECRET=<client-secret>
go test
```
