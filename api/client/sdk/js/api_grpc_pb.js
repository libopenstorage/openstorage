// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var api_pb = require('./api_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_api_annotations_pb = require('./google/api/annotations_pb.js');

function serialize_openstorage_api_ClusterAlertClearRequest(arg) {
  if (!(arg instanceof api_pb.ClusterAlertClearRequest)) {
    throw new Error('Expected argument of type openstorage.api.ClusterAlertClearRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterAlertClearRequest(buffer_arg) {
  return api_pb.ClusterAlertClearRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterAlertClearResponse(arg) {
  if (!(arg instanceof api_pb.ClusterAlertClearResponse)) {
    throw new Error('Expected argument of type openstorage.api.ClusterAlertClearResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterAlertClearResponse(buffer_arg) {
  return api_pb.ClusterAlertClearResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterAlertEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.ClusterAlertEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.ClusterAlertEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterAlertEnumerateRequest(buffer_arg) {
  return api_pb.ClusterAlertEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterAlertEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.ClusterAlertEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.ClusterAlertEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterAlertEnumerateResponse(buffer_arg) {
  return api_pb.ClusterAlertEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterAlertEraseRequest(arg) {
  if (!(arg instanceof api_pb.ClusterAlertEraseRequest)) {
    throw new Error('Expected argument of type openstorage.api.ClusterAlertEraseRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterAlertEraseRequest(buffer_arg) {
  return api_pb.ClusterAlertEraseRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterAlertEraseResponse(arg) {
  if (!(arg instanceof api_pb.ClusterAlertEraseResponse)) {
    throw new Error('Expected argument of type openstorage.api.ClusterAlertEraseResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterAlertEraseResponse(buffer_arg) {
  return api_pb.ClusterAlertEraseResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.ClusterEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.ClusterEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterEnumerateRequest(buffer_arg) {
  return api_pb.ClusterEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.ClusterEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.ClusterEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterEnumerateResponse(buffer_arg) {
  return api_pb.ClusterEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterInspectRequest(arg) {
  if (!(arg instanceof api_pb.ClusterInspectRequest)) {
    throw new Error('Expected argument of type openstorage.api.ClusterInspectRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterInspectRequest(buffer_arg) {
  return api_pb.ClusterInspectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_ClusterInspectResponse(arg) {
  if (!(arg instanceof api_pb.ClusterInspectResponse)) {
    throw new Error('Expected argument of type openstorage.api.ClusterInspectResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_ClusterInspectResponse(buffer_arg) {
  return api_pb.ClusterInspectResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_OpenStorageVolumeCreateRequest(arg) {
  if (!(arg instanceof api_pb.OpenStorageVolumeCreateRequest)) {
    throw new Error('Expected argument of type openstorage.api.OpenStorageVolumeCreateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_OpenStorageVolumeCreateRequest(buffer_arg) {
  return api_pb.OpenStorageVolumeCreateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_OpenStorageVolumeCreateResponse(arg) {
  if (!(arg instanceof api_pb.OpenStorageVolumeCreateResponse)) {
    throw new Error('Expected argument of type openstorage.api.OpenStorageVolumeCreateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_OpenStorageVolumeCreateResponse(buffer_arg) {
  return api_pb.OpenStorageVolumeCreateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeCreateFromVolumeIDRequest(arg) {
  if (!(arg instanceof api_pb.VolumeCreateFromVolumeIDRequest)) {
    throw new Error('Expected argument of type openstorage.api.VolumeCreateFromVolumeIDRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeCreateFromVolumeIDRequest(buffer_arg) {
  return api_pb.VolumeCreateFromVolumeIDRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeCreateFromVolumeIDResponse(arg) {
  if (!(arg instanceof api_pb.VolumeCreateFromVolumeIDResponse)) {
    throw new Error('Expected argument of type openstorage.api.VolumeCreateFromVolumeIDResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeCreateFromVolumeIDResponse(buffer_arg) {
  return api_pb.VolumeCreateFromVolumeIDResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeDeleteRequest(arg) {
  if (!(arg instanceof api_pb.VolumeDeleteRequest)) {
    throw new Error('Expected argument of type openstorage.api.VolumeDeleteRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeDeleteRequest(buffer_arg) {
  return api_pb.VolumeDeleteRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeDeleteResponse(arg) {
  if (!(arg instanceof api_pb.VolumeDeleteResponse)) {
    throw new Error('Expected argument of type openstorage.api.VolumeDeleteResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeDeleteResponse(buffer_arg) {
  return api_pb.VolumeDeleteResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.VolumeEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.VolumeEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeEnumerateRequest(buffer_arg) {
  return api_pb.VolumeEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.VolumeEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.VolumeEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeEnumerateResponse(buffer_arg) {
  return api_pb.VolumeEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeInspectRequest(arg) {
  if (!(arg instanceof api_pb.VolumeInspectRequest)) {
    throw new Error('Expected argument of type openstorage.api.VolumeInspectRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeInspectRequest(buffer_arg) {
  return api_pb.VolumeInspectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeInspectResponse(arg) {
  if (!(arg instanceof api_pb.VolumeInspectResponse)) {
    throw new Error('Expected argument of type openstorage.api.VolumeInspectResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeInspectResponse(buffer_arg) {
  return api_pb.VolumeInspectResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeSnapshotCreateRequest(arg) {
  if (!(arg instanceof api_pb.VolumeSnapshotCreateRequest)) {
    throw new Error('Expected argument of type openstorage.api.VolumeSnapshotCreateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeSnapshotCreateRequest(buffer_arg) {
  return api_pb.VolumeSnapshotCreateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeSnapshotCreateResponse(arg) {
  if (!(arg instanceof api_pb.VolumeSnapshotCreateResponse)) {
    throw new Error('Expected argument of type openstorage.api.VolumeSnapshotCreateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeSnapshotCreateResponse(buffer_arg) {
  return api_pb.VolumeSnapshotCreateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeSnapshotEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.VolumeSnapshotEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.VolumeSnapshotEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeSnapshotEnumerateRequest(buffer_arg) {
  return api_pb.VolumeSnapshotEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeSnapshotEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.VolumeSnapshotEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.VolumeSnapshotEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeSnapshotEnumerateResponse(buffer_arg) {
  return api_pb.VolumeSnapshotEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeSnapshotRestoreRequest(arg) {
  if (!(arg instanceof api_pb.VolumeSnapshotRestoreRequest)) {
    throw new Error('Expected argument of type openstorage.api.VolumeSnapshotRestoreRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeSnapshotRestoreRequest(buffer_arg) {
  return api_pb.VolumeSnapshotRestoreRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_VolumeSnapshotRestoreResponse(arg) {
  if (!(arg instanceof api_pb.VolumeSnapshotRestoreResponse)) {
    throw new Error('Expected argument of type openstorage.api.VolumeSnapshotRestoreResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_VolumeSnapshotRestoreResponse(buffer_arg) {
  return api_pb.VolumeSnapshotRestoreResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var OpenStorageClusterService = exports.OpenStorageClusterService = {
  // Enumerate lists all the nodes in the cluster.
  enumerate: {
    path: '/openstorage.api.OpenStorageCluster/Enumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.ClusterEnumerateRequest,
    responseType: api_pb.ClusterEnumerateResponse,
    requestSerialize: serialize_openstorage_api_ClusterEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_ClusterEnumerateRequest,
    responseSerialize: serialize_openstorage_api_ClusterEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_ClusterEnumerateResponse,
  },
  // Inspect the node given a UUID.
  inspect: {
    path: '/openstorage.api.OpenStorageCluster/Inspect',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.ClusterInspectRequest,
    responseType: api_pb.ClusterInspectResponse,
    requestSerialize: serialize_openstorage_api_ClusterInspectRequest,
    requestDeserialize: deserialize_openstorage_api_ClusterInspectRequest,
    responseSerialize: serialize_openstorage_api_ClusterInspectResponse,
    responseDeserialize: deserialize_openstorage_api_ClusterInspectResponse,
  },
  // Get a list of alerts from the storage cluster
  alertEnumerate: {
    path: '/openstorage.api.OpenStorageCluster/AlertEnumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.ClusterAlertEnumerateRequest,
    responseType: api_pb.ClusterAlertEnumerateResponse,
    requestSerialize: serialize_openstorage_api_ClusterAlertEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_ClusterAlertEnumerateRequest,
    responseSerialize: serialize_openstorage_api_ClusterAlertEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_ClusterAlertEnumerateResponse,
  },
  // Clear the alert for a given resource
  alertClear: {
    path: '/openstorage.api.OpenStorageCluster/AlertClear',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.ClusterAlertClearRequest,
    responseType: api_pb.ClusterAlertClearResponse,
    requestSerialize: serialize_openstorage_api_ClusterAlertClearRequest,
    requestDeserialize: deserialize_openstorage_api_ClusterAlertClearRequest,
    responseSerialize: serialize_openstorage_api_ClusterAlertClearResponse,
    responseDeserialize: deserialize_openstorage_api_ClusterAlertClearResponse,
  },
  // Erases an alert for a given resource
  alertErase: {
    path: '/openstorage.api.OpenStorageCluster/AlertErase',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.ClusterAlertEraseRequest,
    responseType: api_pb.ClusterAlertEraseResponse,
    requestSerialize: serialize_openstorage_api_ClusterAlertEraseRequest,
    requestDeserialize: deserialize_openstorage_api_ClusterAlertEraseRequest,
    responseSerialize: serialize_openstorage_api_ClusterAlertEraseResponse,
    responseDeserialize: deserialize_openstorage_api_ClusterAlertEraseResponse,
  },
};

exports.OpenStorageClusterClient = grpc.makeGenericClientConstructor(OpenStorageClusterService);
var OpenStorageVolumeService = exports.OpenStorageVolumeService = {
  // Creates a new volume
  create: {
    path: '/openstorage.api.OpenStorageVolume/Create',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.OpenStorageVolumeCreateRequest,
    responseType: api_pb.OpenStorageVolumeCreateResponse,
    requestSerialize: serialize_openstorage_api_OpenStorageVolumeCreateRequest,
    requestDeserialize: deserialize_openstorage_api_OpenStorageVolumeCreateRequest,
    responseSerialize: serialize_openstorage_api_OpenStorageVolumeCreateResponse,
    responseDeserialize: deserialize_openstorage_api_OpenStorageVolumeCreateResponse,
  },
  // CreateFromVolumeID creates a new volume cloned from an existing volume
  createFromVolumeID: {
    path: '/openstorage.api.OpenStorageVolume/CreateFromVolumeID',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.VolumeCreateFromVolumeIDRequest,
    responseType: api_pb.VolumeCreateFromVolumeIDResponse,
    requestSerialize: serialize_openstorage_api_VolumeCreateFromVolumeIDRequest,
    requestDeserialize: deserialize_openstorage_api_VolumeCreateFromVolumeIDRequest,
    responseSerialize: serialize_openstorage_api_VolumeCreateFromVolumeIDResponse,
    responseDeserialize: deserialize_openstorage_api_VolumeCreateFromVolumeIDResponse,
  },
  // Delete a volume
  delete: {
    path: '/openstorage.api.OpenStorageVolume/Delete',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.VolumeDeleteRequest,
    responseType: api_pb.VolumeDeleteResponse,
    requestSerialize: serialize_openstorage_api_VolumeDeleteRequest,
    requestDeserialize: deserialize_openstorage_api_VolumeDeleteRequest,
    responseSerialize: serialize_openstorage_api_VolumeDeleteResponse,
    responseDeserialize: deserialize_openstorage_api_VolumeDeleteResponse,
  },
  // Get information on a volume
  inspect: {
    path: '/openstorage.api.OpenStorageVolume/Inspect',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.VolumeInspectRequest,
    responseType: api_pb.VolumeInspectResponse,
    requestSerialize: serialize_openstorage_api_VolumeInspectRequest,
    requestDeserialize: deserialize_openstorage_api_VolumeInspectRequest,
    responseSerialize: serialize_openstorage_api_VolumeInspectResponse,
    responseDeserialize: deserialize_openstorage_api_VolumeInspectResponse,
  },
  // Get a list of volumes
  enumerate: {
    path: '/openstorage.api.OpenStorageVolume/Enumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.VolumeEnumerateRequest,
    responseType: api_pb.VolumeEnumerateResponse,
    requestSerialize: serialize_openstorage_api_VolumeEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_VolumeEnumerateRequest,
    responseSerialize: serialize_openstorage_api_VolumeEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_VolumeEnumerateResponse,
  },
  // Create a snapshot of a volume. This creates an immutable (read-only),
  // point-in-time snapshot of a volume.
  snapshotCreate: {
    path: '/openstorage.api.OpenStorageVolume/SnapshotCreate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.VolumeSnapshotCreateRequest,
    responseType: api_pb.VolumeSnapshotCreateResponse,
    requestSerialize: serialize_openstorage_api_VolumeSnapshotCreateRequest,
    requestDeserialize: deserialize_openstorage_api_VolumeSnapshotCreateRequest,
    responseSerialize: serialize_openstorage_api_VolumeSnapshotCreateResponse,
    responseDeserialize: deserialize_openstorage_api_VolumeSnapshotCreateResponse,
  },
  // Restores a volume to a specified snapshot
  snapshotRestore: {
    path: '/openstorage.api.OpenStorageVolume/SnapshotRestore',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.VolumeSnapshotRestoreRequest,
    responseType: api_pb.VolumeSnapshotRestoreResponse,
    requestSerialize: serialize_openstorage_api_VolumeSnapshotRestoreRequest,
    requestDeserialize: deserialize_openstorage_api_VolumeSnapshotRestoreRequest,
    responseSerialize: serialize_openstorage_api_VolumeSnapshotRestoreResponse,
    responseDeserialize: deserialize_openstorage_api_VolumeSnapshotRestoreResponse,
  },
  // List the number of snapshots for a specific volume
  snapshotEnumerate: {
    path: '/openstorage.api.OpenStorageVolume/SnapshotEnumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.VolumeSnapshotEnumerateRequest,
    responseType: api_pb.VolumeSnapshotEnumerateResponse,
    requestSerialize: serialize_openstorage_api_VolumeSnapshotEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_VolumeSnapshotEnumerateRequest,
    responseSerialize: serialize_openstorage_api_VolumeSnapshotEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_VolumeSnapshotEnumerateResponse,
  },
};

exports.OpenStorageVolumeClient = grpc.makeGenericClientConstructor(OpenStorageVolumeService);
