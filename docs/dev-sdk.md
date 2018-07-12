# SDK Development
This document will highlight how to add new APIs to the OpenStorage SDK.

## SDK Requirements
All SDK APIs and values must satisfy by the following:

* Must be readable
    * SDK APIs and values must be concrete clear values
    * `string` values are for ids or strings. They are not meant to marshal information as a `yaml`, create a concrete _message_ instead.
* No value options passed as `string`
    * Instead of passing "Done", or "paused", use enums for these value. Making it clear to the reader.
* Services:
    * Services should be in the format `OpenStorage<Type>`.
    * Note that the service is a collection of APIs and are grouped as such in the documentation.
        * Here is an example for [OpenStorageClusterService](https://libopenstorage.github.io/w/generated-api.html#openstorageapiopenstoragecluster)
* APIs
    * If it is a new service, then it should have `Create`, `Inspect`, `Delete`, or `Enumerate` style APIs.
    * All APIs **must** have a single message for the request and a single message for the response with the following style: `Sdk<Service Type><Api Name>Request|Response`
* Enums
    * Enum of zero value should be labeled as `XXX_UNKNOWN`, `XXX_NONE`, or `XXX_UNDEFINED` to check if it was not set.
    * All enums must be unique across the entire proto file, not just the single enum.
* Messages
    * Try not to use `uint64`. Instead try to use signed `int64`. (There is a reason for this which is why CSI changed all uint64s to in64s in version 0.2, but I can find out why.)
* Documentation
    * It is imperative that the documentation is correct since it is used to automatically generate the documentation for https://libopenstorage.github.io .The documentation for these values in the proto files can be in Markdown format.

**NOTE** Most importantly is that these APIs _must_ be supported forever. They will almost never be deprecated since at some point we will have many versions of the clients. So please be clear and careful on the API you create.

## Creating new API

### Service
If you are adding a new service, use the following steps:

* Create a new service in the proto file
* Create a new file under `api/server/sdk` with the name `<service>.go`.
* In it create an object which will house the API implementation of the server functions for this service. See [Example](https://github.com/libopenstorage/openstorage/blob/97b0c88d1a9f5517dca4b7d19ce91a0377ebce39/api/server/sdk/cloud_backup.go#L30-L32).
* Initialize this object in [`server.go::New()`](https://github.com/libopenstorage/openstorage/blob/97b0c88d1a9f5517dca4b7d19ce91a0377ebce39/api/server/sdk/server.go#L96-L119)
* Add it the endpoint to the [REST gRPC Gateway](https://github.com/libopenstorage/openstorage/blob/master/api/server/sdk/server.go#L202).

### API
To add an API, follow the following steps:

* Create a new API in a service proto file and create its messages.
    * It is **HIGHLY** recommended that you have these messages reviewed _first_ before sending your PR. The easiest way is to create an [Issue](https://github.com/libopenstorage/openstorage/issues/new) with the description of the plan and the proto file API and messages. If not you may have to change all your code in case their is a suggestion on changes to your proto file.
* Generate the Golang bindings by running: `make docker-proto`.
* Add the implementation of the API server interface to the appropriate service file in `api/server/sdk`. You are also welcomed to create new files in that directory which are prefixed by the service name, here is an example: [volume_node_ops.go](https://github.com/libopenstorage/openstorage/blob/master/api/server/sdk/volume_node_ops.go)
* The implementation should only communicate with the OpenStorage golang interfaces, never the REST API.
* You _must_ provide unit tests for your changes which utilize either a mock cluster or a mock driver.
* APIs must check for the required parameters in the message and unit tests must confirm these checks.
* If your test is not supported by the [`fake`](https://github.com/libopenstorage/openstorage/blob/master/volume/drivers/fake/fake.go) driver, please add support for it. It is essential that the `fake` driver supports the your API since it will be used by developers to write their clients using the docker container [as shown in the documentation](https://libopenstorage.github.io/w/#quick-example).

### Functional Testing
To do a functional test using the `fake` driver, do the following:

* Type: `make launch-sdk`
    * This will create a new container with the SDK and run it on your system. Note, if you have one already running, you must stop that container before running this command.
* Use a browser to execute your command.
    * Go to http://127.0.0.1:9110/swagger-ui then click on the command you want to try, then click on `Try it now`.
    * Change or adjust the input request as needed, then click on the `Execute` command.
    * Inspect the response from the server.
    
## Dealing with conflicts on generated files
When rebasing files you may get conflicts on generated files. If you do, just accept the incoming generated files (referred by git as `--ours`) then once all the rebases are done, regenerate again, and commit.

Here are the commands you may need:

```
$ git rebase master
<-- Conflicts. For each conflict on a generated file, repeat: -->
$ git checkout --ours <file with conflict>
$ git add <file with conflict>
$ git rebase --continue
```

