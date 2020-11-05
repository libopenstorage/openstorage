# SDK Development Guide
This document will highlight how to add new APIs to the OpenStorage SDK.

## OpenStorage SDK Style Guide

### Protocol Buffer Style Guide

The SDK follows the [Protocol Buffer Style Guide](https://developers.google.com/protocol-buffers/docs/style). From the guide:

>Note that protocol buffer style has evolved over time, so it is likely that you will see .proto files written in different conventions or styles. Please respect the existing style when you modify these files. Consistency is key. However, it is best to adopt the current best style when you are creating a new .proto file.

Therefore, for any new messages, enums, etc, please follow the style guide.

### OpenStorage Style

All SDK APIs and values must satisfy by the following:

#### Version

Any changes to the protocol must bump the version by one. On the _master_ branch, the _minor_ number is bumped. On _release_ branches, the _patch_ number is bumped.

#### Readability 

**APIs Must be readable**. SDK APIs and values must be concrete clear values. They are used not just by Portworx, but also non-Portworx developers which do not have an understanding of the internals of the Portworx cluster.

#### `string` values.

* `string` types should be used only for ids, messages, or opaque values. They are not meant to marshal information as a `yaml`. Instead create a concrete _message_.
* Only use `map<string, string>` for opaque values like labels, key-value pairs, etc. Do not use them for operations. Use enums instead.
* Value options should not be passed as `string`. Instead of passing "Done", or "paused", use enums for these value, making it clear to the reader.

#### SDK Services

* Services contain RPC functions
* Services should be in the format `OpenStorage<Service Name>`.
* Note that the service is a collection of APIs and are grouped as such in the documentation.
    * Here is an example for [OpenStorageClusterService](https://libopenstorage.github.io/w/release-7.0.generated-api.html#serviceopenstorageapiopenstoragevolume)

##### SDK APIs (RPCs)

* If it is a new service, then it should have `Create`, `Inspect`, `Delete`, or `Enumerate` style APIs, if possible.
* All APIs **must** have a single message for the request and a single message for the response with the following style: `Sdk<Service Type><Api Name>Request|Response`
* RPCs will be created as _methods_ to the service _object_, therefore there is
  no need to add the service name as part of the RPC. For example,
  use `Foo`, or `Bar` instead or `ServiceFoo` or `ServiceBar` as RPC names.

##### Enums

* Follow the [Google protobuf style for enums](https://developers.google.com/protocol-buffers/docs/style#enums)
* Enum of zero value should be labeled as `XXX_UNKNOWN`, `XXX_NONE`, or `XXX_UNDEFINED` to check if it was not set.
* Wrap enums in messages so that their string values are clearer. Example:

```proto
// Xattr defines implementation specific volume attribute
message Xattr {
  enum Value {
    // Value is uninitialized or unknown
    UNSPECIFIED = 0;
    // Enable on-demand copy-on-write on the volume
    COW_ON_DEMAND = 1;
  }
}
```

Wrapping an enum in a message also has the benefit of not needing to prefix the enums with namespaced information. For example, instead of using the enum `XATTR_UNSPECIFIED`, the example above uses just `UNSPECIFIED` since it is inide the `Xattr` message. The generated code will be namepaced:

```go
type Xattr_Value int32

const (
	// Value is uninitialized or unknown
	Xattr_UNSPECIFIED Xattr_Value = 0
	// Enable on-demand copy-on-write on the volume
	Xattr_COW_ON_DEMAND Xattr_Value = 1
)

var Xattr_Value_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "COW_ON_DEMAND",
}
var Xattr_Value_value = map[string]int32{
	"UNSPECIFIED":   0,
	"COW_ON_DEMAND": 1,
}
```

##### Security

OpenStorage has a number of default roles used to authorize access to an API.
Please make sure your API is accessible by the appropriate role. For example,
the role `system.admin` has access to all APIs while the role `system.user`
only has access to volume management APIs.

The default roles are configured under [`pkg/roles`](https://github.com/libopenstorage/openstorage/blob/9e8fcaedd4e0464772fbe71fd75a37515c79c910/pkg/role/sdkserviceapi.go#L53). See [SdkRule](https://github.com/libopenstorage/openstorage/blob/9e8fcaedd4e0464772fbe71fd75a37515c79c910/api/api.proto#L4301) for more information.

If you have any questions, please do not hesitate to ask.

#### Messages

* Try not to use `uint64`. Instead try to use signed `int64`. (There is a reason for this which is why CSI changed all uint64s to in64s in version 0.2, but I can find out why. I think it has to do with Java gRPC)

##### Field Numbers

* If it is a new message, start with the field number of `1`.
* If it is an addition to a message, continue the field number sequence by one.
* If you are using `oneof` you may want to start with a large value for the
  field number so that they do not interfere with other values in the message:

```proto
  string s3_storage_class = 7;

  // Start at field number 200 to allow for expansion
  oneof credential_type {
    // Credentials for AWS/S3
    SdkAwsCredentialRequest aws_credential = 200;
    // Credentials for Azure
    SdkAzureCredentialRequest azure_credential = 201;
    // Credentials for Google
    SdkGoogleCredentialRequest google_credential = 202;
  }
```

#### REST

REST endpoints are autogenerated from the protofile by the grpc-gateway protoc compiler. All OpenStorage SDK APIs should add the appropriate information to generate a REST endpoint for the service. Here is an example:

```proto
  rpc Inspect(SdkRoleInspectRequest)
    returns (SdkRoleInspectResponse){
      option(google.api.http) = {
        get: "/v1/roles/inspect/{name}"
      };
    }

  // Delete an existing role
  rpc Delete(SdkRoleDeleteRequest)
    returns (SdkRoleDeleteResponse){
      option(google.api.http) = {
        delete: "/v1/roles/{name}"
      };
    }

  // Update an existing role
  rpc Update(SdkRoleUpdateRequest)
    returns (SdkRoleUpdateResponse){
      option(google.api.http) = {
        put: "/v1/roles"
        body: "*"
      };
    }
```

Here are the guidelines for REST in OpenStorage SDK:

* Endpoint must be prefixed as follows: `/v1/<service name>/<rpc name if needed>/{any variables if needed}`.
* Use the appropriate HTTP method. Here are some guidelines:
    * For _Create_ RPCs use the `post` http method
    * For _Inspect_ RPCs use the `get` http method
    * For _Update_ RPCs use the `put` http method
    * For _Delete_ RPCs use the `delete` http method
* Use `get` for non-mutable calls.
* Use `put` with `body: "*"` most calls that need to send a message to the SDK server.

#### Documentation

* It is _imperative_ that the comments are correct since they are used to automatically generate the documentation for https://libopenstorage.github.io . The documentation for these values in the proto files can be in Markdown format.

* **Documenting Messages**
    * Document each value of the message.
    * Do not use Golang style. Do not repeat the name of the variable in _Golang Camel Format_ in the comment to document it since the variable could be in other styles in other languages. For example:

```proto
// Provides volume's exclusive bytes and its total usage. This cannot be
// retrieved individually and is obtained as part node's usage for a given
// node.
message VolumeUsage {
   // id for the volume/snapshot
  string volume_id = 1;
  // name of the volume/snapshot
  string volume_name = 2;
  // uuid of the pool that this volume belongs to
  string pool_uuid = 3;
  // size in bytes exclusively used by the volume/snapshot
  uint64 exclusive_bytes = 4;
  //  size in bytes by the volume/snapshot
  uint64 total_bytes = 5;
  // set to true if this volume is snapshot created by cloudbackups
  bool local_cloud_snapshot = 6;
}
```

#### Deprecation

* **NOTE:** Most importantly is that these APIs _must_ be supported forever once released. 
* They will almost never be deprecated since at some point we will have many versions of the clients. So please be clear and careful on the API you create.
* If we need to change or update, you can always **add**.

Here is the process if you would like to deprecate:

1. According to [proto3 Language Guide](https://developers.google.com/protocol-buffers/docs/proto3) set the value in the message to deprecated and add a `(deprecated)` string to the comment as follows:

```proto
// (deprecated) Field documentation here
int32 field = 6 [deprecated = true];
```

2. Comment in the SDK_CHANGELOG that the value is deprecated.
3. Provide at least two releases before removing support for that value in the message. Make sure to document in the release notes of the product the deprecation.
4. Once at least two releases have passed. Reserve the field number as shown in the [proto3 Language Guide](https://developers.google.com/protocol-buffers/docs/proto3#reserved):

```proto
message Foo {
  reserved 6;
}
```

It is essential that no values override the field number when updating or replacing. From the guide:

> Note: If you update a message type by entirely removing a field, or commenting it out, future users can reuse the field number when making their own updates to the type. This can cause severe issues if they later load old versions of the same .proto, including data corruption, privacy bugs, and so on.

## Creating new API

### Service

If you are adding a new service, use the following steps:

* Create a new service in the proto file
* For Volume services:
    * Create a new file under `api/server/sdk` with the name `<service>.go`.
    * In it create an object which will house the API implementation of the server functions for this service. See [Example](https://github.com/libopenstorage/openstorage/blob/97b0c88d1a9f5517dca4b7d19ce91a0377ebce39/api/server/sdk/cloud_backup.go#L30-L32).
    * Initialize this object in [`server.go::New()`](https://github.com/libopenstorage/openstorage/blob/97b0c88d1a9f5517dca4b7d19ce91a0377ebce39/api/server/sdk/server.go#L96-L119)
    * Add it the endpoint to the [REST gRPC Gateway](https://github.com/libopenstorage/openstorage/blob/master/api/server/sdk/server.go#L202).
* For non-Volume services:
    * Create a Golang implementation of the service in `pkg/`
    * Implement the gRPC generated server definition there
    * Write unit tests for your implementation if possible
    * See [Role Manager](https://github.com/libopenstorage/openstorage/tree/master/pkg/role) as an example.
    * Adding your service to the SDK server:
        * NOTE: We are moving away from adding services directly in the SDK `server.go`. Instead, instantiate your object, then add it to the SDK `ServerConfig{}`. This can be done by adding services to `ServerConfig.GrpcServerExtensions[]` and generated REST services to `ServerConfig.RestServerExtensions[]`. For more information see [`ServerConfig{}`](https://github.com/libopenstorage/openstorage/blob/9e8fcaedd4e0464772fbe71fd75a37515c79c910/api/server/sdk/server.go#L116-L135).

### API
To add an API, follow the following steps:

* Create a new API in a service proto file and create its messages.
    * It is **HIGHLY** recommended that you have these messages reviewed _first_ before sending your PR. The easiest way is to create an [Issue](https://github.com/libopenstorage/openstorage/issues/new) with the description of the plan and the proto file API and messages. If not you may have to change all your code in case their is a suggestion on changes to your proto file.
* Generate the Golang bindings by running: `make docker-proto`.
* Add the implementation as described above
* The implementation should only communicate with the OpenStorage golang interfaces, never the REST API.
* APIs must check for the required parameters in the message and unit tests must confirm these checks.
* If your API is not supported by the [`fake`](https://github.com/libopenstorage/openstorage/blob/master/volume/drivers/fake/fake.go) driver, please add support for it. It is essential that the `fake` driver supports the your API since it will be used by developers to write their clients using the docker container [as shown in the documentation](https://libopenstorage.github.io/w/#quick-example).

### Development Testing
To do a development testing using the `fake` driver, do the following:

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

# Release Management
The following steps can be used to publish it once the SDK is ready for a new release and the version number has been updated:

* Update [sdk-test](https://github.com/libopenstorage/sdk-test) to test the new functionality.
* Update [docs](https://github.com/libopenstorage/libopenstorage.github.io) _Reference_ and _Changelog_.
* Update [openstorage-sdk-clients](https://github.com/libopenstorage/openstorage-sdk-clients) to regenerate new clients.


