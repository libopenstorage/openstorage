// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var api_pb = require('./api_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

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
};

exports.OpenStorageClusterClient = grpc.makeGenericClientConstructor(OpenStorageClusterService);
