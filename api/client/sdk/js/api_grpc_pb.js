// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var api_pb = require('./api_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_api_annotations_pb = require('./google/api/annotations_pb.js');

function serialize_openstorage_api_SdkClusterAlertClearRequest(arg) {
  if (!(arg instanceof api_pb.SdkClusterAlertClearRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterAlertClearRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterAlertClearRequest(buffer_arg) {
  return api_pb.SdkClusterAlertClearRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterAlertClearResponse(arg) {
  if (!(arg instanceof api_pb.SdkClusterAlertClearResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterAlertClearResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterAlertClearResponse(buffer_arg) {
  return api_pb.SdkClusterAlertClearResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterAlertEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.SdkClusterAlertEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterAlertEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterAlertEnumerateRequest(buffer_arg) {
  return api_pb.SdkClusterAlertEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterAlertEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.SdkClusterAlertEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterAlertEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterAlertEnumerateResponse(buffer_arg) {
  return api_pb.SdkClusterAlertEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterAlertEraseRequest(arg) {
  if (!(arg instanceof api_pb.SdkClusterAlertEraseRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterAlertEraseRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterAlertEraseRequest(buffer_arg) {
  return api_pb.SdkClusterAlertEraseRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterAlertEraseResponse(arg) {
  if (!(arg instanceof api_pb.SdkClusterAlertEraseResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterAlertEraseResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterAlertEraseResponse(buffer_arg) {
  return api_pb.SdkClusterAlertEraseResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.SdkClusterEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterEnumerateRequest(buffer_arg) {
  return api_pb.SdkClusterEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.SdkClusterEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterEnumerateResponse(buffer_arg) {
  return api_pb.SdkClusterEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterInspectRequest(arg) {
  if (!(arg instanceof api_pb.SdkClusterInspectRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterInspectRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterInspectRequest(buffer_arg) {
  return api_pb.SdkClusterInspectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterInspectResponse(arg) {
  if (!(arg instanceof api_pb.SdkClusterInspectResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterInspectResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterInspectResponse(buffer_arg) {
  return api_pb.SdkClusterInspectResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialCreateAWSRequest(arg) {
  if (!(arg instanceof api_pb.SdkCredentialCreateAWSRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialCreateAWSRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialCreateAWSRequest(buffer_arg) {
  return api_pb.SdkCredentialCreateAWSRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialCreateAWSResponse(arg) {
  if (!(arg instanceof api_pb.SdkCredentialCreateAWSResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialCreateAWSResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialCreateAWSResponse(buffer_arg) {
  return api_pb.SdkCredentialCreateAWSResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialCreateAzureRequest(arg) {
  if (!(arg instanceof api_pb.SdkCredentialCreateAzureRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialCreateAzureRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialCreateAzureRequest(buffer_arg) {
  return api_pb.SdkCredentialCreateAzureRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialCreateAzureResponse(arg) {
  if (!(arg instanceof api_pb.SdkCredentialCreateAzureResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialCreateAzureResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialCreateAzureResponse(buffer_arg) {
  return api_pb.SdkCredentialCreateAzureResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialCreateGoogleRequest(arg) {
  if (!(arg instanceof api_pb.SdkCredentialCreateGoogleRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialCreateGoogleRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialCreateGoogleRequest(buffer_arg) {
  return api_pb.SdkCredentialCreateGoogleRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialCreateGoogleResponse(arg) {
  if (!(arg instanceof api_pb.SdkCredentialCreateGoogleResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialCreateGoogleResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialCreateGoogleResponse(buffer_arg) {
  return api_pb.SdkCredentialCreateGoogleResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialDeleteRequest(arg) {
  if (!(arg instanceof api_pb.SdkCredentialDeleteRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialDeleteRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialDeleteRequest(buffer_arg) {
  return api_pb.SdkCredentialDeleteRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialDeleteResponse(arg) {
  if (!(arg instanceof api_pb.SdkCredentialDeleteResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialDeleteResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialDeleteResponse(buffer_arg) {
  return api_pb.SdkCredentialDeleteResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialEnumerateAWSRequest(arg) {
  if (!(arg instanceof api_pb.SdkCredentialEnumerateAWSRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialEnumerateAWSRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialEnumerateAWSRequest(buffer_arg) {
  return api_pb.SdkCredentialEnumerateAWSRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialEnumerateAWSResponse(arg) {
  if (!(arg instanceof api_pb.SdkCredentialEnumerateAWSResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialEnumerateAWSResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialEnumerateAWSResponse(buffer_arg) {
  return api_pb.SdkCredentialEnumerateAWSResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialEnumerateAzureRequest(arg) {
  if (!(arg instanceof api_pb.SdkCredentialEnumerateAzureRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialEnumerateAzureRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialEnumerateAzureRequest(buffer_arg) {
  return api_pb.SdkCredentialEnumerateAzureRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialEnumerateAzureResponse(arg) {
  if (!(arg instanceof api_pb.SdkCredentialEnumerateAzureResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialEnumerateAzureResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialEnumerateAzureResponse(buffer_arg) {
  return api_pb.SdkCredentialEnumerateAzureResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialEnumerateGoogleRequest(arg) {
  if (!(arg instanceof api_pb.SdkCredentialEnumerateGoogleRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialEnumerateGoogleRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialEnumerateGoogleRequest(buffer_arg) {
  return api_pb.SdkCredentialEnumerateGoogleRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialEnumerateGoogleResponse(arg) {
  if (!(arg instanceof api_pb.SdkCredentialEnumerateGoogleResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialEnumerateGoogleResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialEnumerateGoogleResponse(buffer_arg) {
  return api_pb.SdkCredentialEnumerateGoogleResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialValidateRequest(arg) {
  if (!(arg instanceof api_pb.SdkCredentialValidateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialValidateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialValidateRequest(buffer_arg) {
  return api_pb.SdkCredentialValidateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCredentialValidateResponse(arg) {
  if (!(arg instanceof api_pb.SdkCredentialValidateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCredentialValidateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCredentialValidateResponse(buffer_arg) {
  return api_pb.SdkCredentialValidateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeAttachRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeAttachRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeAttachRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeAttachRequest(buffer_arg) {
  return api_pb.SdkVolumeAttachRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeAttachResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeAttachResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeAttachResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeAttachResponse(buffer_arg) {
  return api_pb.SdkVolumeAttachResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeCreateFromVolumeIdRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeCreateFromVolumeIdRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeCreateFromVolumeIdRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeCreateFromVolumeIdRequest(buffer_arg) {
  return api_pb.SdkVolumeCreateFromVolumeIdRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeCreateFromVolumeIdResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeCreateFromVolumeIdResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeCreateFromVolumeIdResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeCreateFromVolumeIdResponse(buffer_arg) {
  return api_pb.SdkVolumeCreateFromVolumeIdResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeCreateRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeCreateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeCreateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeCreateRequest(buffer_arg) {
  return api_pb.SdkVolumeCreateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeCreateResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeCreateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeCreateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeCreateResponse(buffer_arg) {
  return api_pb.SdkVolumeCreateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeDeleteRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeDeleteRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeDeleteRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeDeleteRequest(buffer_arg) {
  return api_pb.SdkVolumeDeleteRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeDeleteResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeDeleteResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeDeleteResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeDeleteResponse(buffer_arg) {
  return api_pb.SdkVolumeDeleteResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeDetachRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeDetachRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeDetachRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeDetachRequest(buffer_arg) {
  return api_pb.SdkVolumeDetachRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeDetachResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeDetachResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeDetachResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeDetachResponse(buffer_arg) {
  return api_pb.SdkVolumeDetachResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeEnumerateRequest(buffer_arg) {
  return api_pb.SdkVolumeEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeEnumerateResponse(buffer_arg) {
  return api_pb.SdkVolumeEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeInspectRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeInspectRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeInspectRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeInspectRequest(buffer_arg) {
  return api_pb.SdkVolumeInspectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeInspectResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeInspectResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeInspectResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeInspectResponse(buffer_arg) {
  return api_pb.SdkVolumeInspectResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeMountRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeMountRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeMountRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeMountRequest(buffer_arg) {
  return api_pb.SdkVolumeMountRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeMountResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeMountResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeMountResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeMountResponse(buffer_arg) {
  return api_pb.SdkVolumeMountResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeSnapshotCreateRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeSnapshotCreateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeSnapshotCreateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeSnapshotCreateRequest(buffer_arg) {
  return api_pb.SdkVolumeSnapshotCreateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeSnapshotCreateResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeSnapshotCreateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeSnapshotCreateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeSnapshotCreateResponse(buffer_arg) {
  return api_pb.SdkVolumeSnapshotCreateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeSnapshotEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeSnapshotEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeSnapshotEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeSnapshotEnumerateRequest(buffer_arg) {
  return api_pb.SdkVolumeSnapshotEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeSnapshotEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeSnapshotEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeSnapshotEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeSnapshotEnumerateResponse(buffer_arg) {
  return api_pb.SdkVolumeSnapshotEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeSnapshotRestoreRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeSnapshotRestoreRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeSnapshotRestoreRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeSnapshotRestoreRequest(buffer_arg) {
  return api_pb.SdkVolumeSnapshotRestoreRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeSnapshotRestoreResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeSnapshotRestoreResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeSnapshotRestoreResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeSnapshotRestoreResponse(buffer_arg) {
  return api_pb.SdkVolumeSnapshotRestoreResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeUnmountRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeUnmountRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeUnmountRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeUnmountRequest(buffer_arg) {
  return api_pb.SdkVolumeUnmountRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeUnmountResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeUnmountResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeUnmountResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeUnmountResponse(buffer_arg) {
  return api_pb.SdkVolumeUnmountResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var OpenStorageClusterService = exports.OpenStorageClusterService = {
  // Enumerate lists all the nodes in the cluster.
  enumerate: {
    path: '/openstorage.api.OpenStorageCluster/Enumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkClusterEnumerateRequest,
    responseType: api_pb.SdkClusterEnumerateResponse,
    requestSerialize: serialize_openstorage_api_SdkClusterEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkClusterEnumerateRequest,
    responseSerialize: serialize_openstorage_api_SdkClusterEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkClusterEnumerateResponse,
  },
  // Inspect the node given a UUID.
  inspect: {
    path: '/openstorage.api.OpenStorageCluster/Inspect',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkClusterInspectRequest,
    responseType: api_pb.SdkClusterInspectResponse,
    requestSerialize: serialize_openstorage_api_SdkClusterInspectRequest,
    requestDeserialize: deserialize_openstorage_api_SdkClusterInspectRequest,
    responseSerialize: serialize_openstorage_api_SdkClusterInspectResponse,
    responseDeserialize: deserialize_openstorage_api_SdkClusterInspectResponse,
  },
  // Get a list of alerts from the storage cluster
  alertEnumerate: {
    path: '/openstorage.api.OpenStorageCluster/AlertEnumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkClusterAlertEnumerateRequest,
    responseType: api_pb.SdkClusterAlertEnumerateResponse,
    requestSerialize: serialize_openstorage_api_SdkClusterAlertEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkClusterAlertEnumerateRequest,
    responseSerialize: serialize_openstorage_api_SdkClusterAlertEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkClusterAlertEnumerateResponse,
  },
  // Clear the alert for a given resource
  alertClear: {
    path: '/openstorage.api.OpenStorageCluster/AlertClear',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkClusterAlertClearRequest,
    responseType: api_pb.SdkClusterAlertClearResponse,
    requestSerialize: serialize_openstorage_api_SdkClusterAlertClearRequest,
    requestDeserialize: deserialize_openstorage_api_SdkClusterAlertClearRequest,
    responseSerialize: serialize_openstorage_api_SdkClusterAlertClearResponse,
    responseDeserialize: deserialize_openstorage_api_SdkClusterAlertClearResponse,
  },
  // Erases an alert for a given resource
  alertErase: {
    path: '/openstorage.api.OpenStorageCluster/AlertErase',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkClusterAlertEraseRequest,
    responseType: api_pb.SdkClusterAlertEraseResponse,
    requestSerialize: serialize_openstorage_api_SdkClusterAlertEraseRequest,
    requestDeserialize: deserialize_openstorage_api_SdkClusterAlertEraseRequest,
    responseSerialize: serialize_openstorage_api_SdkClusterAlertEraseResponse,
    responseDeserialize: deserialize_openstorage_api_SdkClusterAlertEraseResponse,
  },
};

exports.OpenStorageClusterClient = grpc.makeGenericClientConstructor(OpenStorageClusterService);
var OpenStorageVolumeService = exports.OpenStorageVolumeService = {
  // Creates a new volume
  create: {
    path: '/openstorage.api.OpenStorageVolume/Create',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeCreateRequest,
    responseType: api_pb.SdkVolumeCreateResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeCreateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeCreateRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeCreateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeCreateResponse,
  },
  // CreateFromVolumeId creates a new volume cloned from an existing volume
  createFromVolumeId: {
    path: '/openstorage.api.OpenStorageVolume/CreateFromVolumeId',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeCreateFromVolumeIdRequest,
    responseType: api_pb.SdkVolumeCreateFromVolumeIdResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeCreateFromVolumeIdRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeCreateFromVolumeIdRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeCreateFromVolumeIdResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeCreateFromVolumeIdResponse,
  },
  // Delete a volume
  delete: {
    path: '/openstorage.api.OpenStorageVolume/Delete',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeDeleteRequest,
    responseType: api_pb.SdkVolumeDeleteResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeDeleteRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeDeleteRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeDeleteResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeDeleteResponse,
  },
  // Get information on a volume
  inspect: {
    path: '/openstorage.api.OpenStorageVolume/Inspect',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeInspectRequest,
    responseType: api_pb.SdkVolumeInspectResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeInspectRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeInspectRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeInspectResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeInspectResponse,
  },
  // Get a list of volumes
  enumerate: {
    path: '/openstorage.api.OpenStorageVolume/Enumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeEnumerateRequest,
    responseType: api_pb.SdkVolumeEnumerateResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeEnumerateRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeEnumerateResponse,
  },
  // Create a snapshot of a volume. This creates an immutable (read-only),
  // point-in-time snapshot of a volume.
  snapshotCreate: {
    path: '/openstorage.api.OpenStorageVolume/SnapshotCreate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeSnapshotCreateRequest,
    responseType: api_pb.SdkVolumeSnapshotCreateResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeSnapshotCreateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeSnapshotCreateRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeSnapshotCreateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeSnapshotCreateResponse,
  },
  // Restores a volume to a specified snapshot
  snapshotRestore: {
    path: '/openstorage.api.OpenStorageVolume/SnapshotRestore',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeSnapshotRestoreRequest,
    responseType: api_pb.SdkVolumeSnapshotRestoreResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeSnapshotRestoreRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeSnapshotRestoreRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeSnapshotRestoreResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeSnapshotRestoreResponse,
  },
  // List the number of snapshots for a specific volume
  snapshotEnumerate: {
    path: '/openstorage.api.OpenStorageVolume/SnapshotEnumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeSnapshotEnumerateRequest,
    responseType: api_pb.SdkVolumeSnapshotEnumerateResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeSnapshotEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeSnapshotEnumerateRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeSnapshotEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeSnapshotEnumerateResponse,
  },
  // Attach device to host
  attach: {
    path: '/openstorage.api.OpenStorageVolume/Attach',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeAttachRequest,
    responseType: api_pb.SdkVolumeAttachResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeAttachRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeAttachRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeAttachResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeAttachResponse,
  },
  // Detaches the volume from the node.
  detach: {
    path: '/openstorage.api.OpenStorageVolume/Detach',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeDetachRequest,
    responseType: api_pb.SdkVolumeDetachResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeDetachRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeDetachRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeDetachResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeDetachResponse,
  },
  // Attaches the volume to a node.
  mount: {
    path: '/openstorage.api.OpenStorageVolume/Mount',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeMountRequest,
    responseType: api_pb.SdkVolumeMountResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeMountRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeMountRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeMountResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeMountResponse,
  },
  // Unmount volume at specified path
  unmount: {
    path: '/openstorage.api.OpenStorageVolume/Unmount',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeUnmountRequest,
    responseType: api_pb.SdkVolumeUnmountResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeUnmountRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeUnmountRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeUnmountResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeUnmountResponse,
  },
};

exports.OpenStorageVolumeClient = grpc.makeGenericClientConstructor(OpenStorageVolumeService);
var OpenStorageCredentialsService = exports.OpenStorageCredentialsService = {
  // Provide credentials to OpenStorage and if valid,
  // it will return an identifier to the credentials
  //
  // Create credential for AWS S3 and if valid ,
  // returns a unique identifier
  createForAWS: {
    path: '/openstorage.api.OpenStorageCredentials/CreateForAWS',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCredentialCreateAWSRequest,
    responseType: api_pb.SdkCredentialCreateAWSResponse,
    requestSerialize: serialize_openstorage_api_SdkCredentialCreateAWSRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCredentialCreateAWSRequest,
    responseSerialize: serialize_openstorage_api_SdkCredentialCreateAWSResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCredentialCreateAWSResponse,
  },
  // Create credential for Azure and if valid ,
  // returns a unique identifier
  createForAzure: {
    path: '/openstorage.api.OpenStorageCredentials/CreateForAzure',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCredentialCreateAzureRequest,
    responseType: api_pb.SdkCredentialCreateAzureResponse,
    requestSerialize: serialize_openstorage_api_SdkCredentialCreateAzureRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCredentialCreateAzureRequest,
    responseSerialize: serialize_openstorage_api_SdkCredentialCreateAzureResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCredentialCreateAzureResponse,
  },
  // Create credential for Google and if valid ,
  // returns a unique identifier
  createForGoogle: {
    path: '/openstorage.api.OpenStorageCredentials/CreateForGoogle',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCredentialCreateGoogleRequest,
    responseType: api_pb.SdkCredentialCreateGoogleResponse,
    requestSerialize: serialize_openstorage_api_SdkCredentialCreateGoogleRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCredentialCreateGoogleRequest,
    responseSerialize: serialize_openstorage_api_SdkCredentialCreateGoogleResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCredentialCreateGoogleResponse,
  },
  // EnumerateForAWS lists the configured AWS credentials
  enumerateForAWS: {
    path: '/openstorage.api.OpenStorageCredentials/EnumerateForAWS',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCredentialEnumerateAWSRequest,
    responseType: api_pb.SdkCredentialEnumerateAWSResponse,
    requestSerialize: serialize_openstorage_api_SdkCredentialEnumerateAWSRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCredentialEnumerateAWSRequest,
    responseSerialize: serialize_openstorage_api_SdkCredentialEnumerateAWSResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCredentialEnumerateAWSResponse,
  },
  // EnumerateForAzure lists the configured Azure credentials
  enumerateForAzure: {
    path: '/openstorage.api.OpenStorageCredentials/EnumerateForAzure',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCredentialEnumerateAzureRequest,
    responseType: api_pb.SdkCredentialEnumerateAzureResponse,
    requestSerialize: serialize_openstorage_api_SdkCredentialEnumerateAzureRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCredentialEnumerateAzureRequest,
    responseSerialize: serialize_openstorage_api_SdkCredentialEnumerateAzureResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCredentialEnumerateAzureResponse,
  },
  // EnumerateForGoogle lists the configured Google credentials
  enumerateForGoogle: {
    path: '/openstorage.api.OpenStorageCredentials/EnumerateForGoogle',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCredentialEnumerateGoogleRequest,
    responseType: api_pb.SdkCredentialEnumerateGoogleResponse,
    requestSerialize: serialize_openstorage_api_SdkCredentialEnumerateGoogleRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCredentialEnumerateGoogleRequest,
    responseSerialize: serialize_openstorage_api_SdkCredentialEnumerateGoogleResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCredentialEnumerateGoogleResponse,
  },
  // Delete a specified credential
  credentialDelete: {
    path: '/openstorage.api.OpenStorageCredentials/CredentialDelete',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCredentialDeleteRequest,
    responseType: api_pb.SdkCredentialDeleteResponse,
    requestSerialize: serialize_openstorage_api_SdkCredentialDeleteRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCredentialDeleteRequest,
    responseSerialize: serialize_openstorage_api_SdkCredentialDeleteResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCredentialDeleteResponse,
  },
  // Validate a specified credential
  credentialValidate: {
    path: '/openstorage.api.OpenStorageCredentials/CredentialValidate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCredentialValidateRequest,
    responseType: api_pb.SdkCredentialValidateResponse,
    requestSerialize: serialize_openstorage_api_SdkCredentialValidateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCredentialValidateRequest,
    responseSerialize: serialize_openstorage_api_SdkCredentialValidateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCredentialValidateResponse,
  },
};

exports.OpenStorageCredentialsClient = grpc.makeGenericClientConstructor(OpenStorageCredentialsService);
