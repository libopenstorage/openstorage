
### To test GCP

You will first need to create a GCP instance and then provide details of this instance as below.

```bash
export GOOGLE_APPLICATION_CREDENTIALS=<path-to-service-account-json-file>
export GCE_INSTANCE_NAME=<gce-instance-name>
export GCE_INSTANCE_ZONE=<gce-instance-zone>
export GCE_INSTANCE_PROJECT=<gce-project-name>

go test -v
```

#### To get path-to-service-account-json-file
* Go to GCP console -> Compute Engine -> VM Instances -> `<your-instance>`
* Find the service account used for this instance in the "Service Account" section
* Go to GCP console -> IAM & admin -> Service Accounts -> `<instance-service-account>`
* Select "Create key" and download the .json file
* Set GOOGLE_APPLICATION_CREDENTIALS to the path of the .json file


