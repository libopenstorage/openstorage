### To test Azure

You will first need to create a Azure instance and then provide details of this instance as below.

```bash
export AZURE_INSTANCE_NAME=<instance-name>
export AZURE_SUBSCRIPTION_ID=<subscription-id>
export AZURE_RESOURCE_GROUP_NAME=<resource-group-name-of-instance>
export AZURE_ENVIRONMENT=<azure-cloud-environment>
export AZURE_TENANT_ID=<tenant-id>
export AZURE_CLIENT_ID=<client-id>
export AZURE_CLIENT_SECRET=<client-secret>
go test
```
