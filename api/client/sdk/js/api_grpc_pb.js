// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var api_pb = require('./api_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_api_annotations_pb = require('./google/api/annotations_pb.js');

function serialize_openstorage_api_SdkCloudBackupCatalogRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupCatalogRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupCatalogRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupCatalogRequest(buffer_arg) {
  return api_pb.SdkCloudBackupCatalogRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupCatalogResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupCatalogResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupCatalogResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupCatalogResponse(buffer_arg) {
  return api_pb.SdkCloudBackupCatalogResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupCreateRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupCreateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupCreateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupCreateRequest(buffer_arg) {
  return api_pb.SdkCloudBackupCreateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupCreateResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupCreateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupCreateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupCreateResponse(buffer_arg) {
  return api_pb.SdkCloudBackupCreateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupDeleteAllRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupDeleteAllRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupDeleteAllRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupDeleteAllRequest(buffer_arg) {
  return api_pb.SdkCloudBackupDeleteAllRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupDeleteAllResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupDeleteAllResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupDeleteAllResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupDeleteAllResponse(buffer_arg) {
  return api_pb.SdkCloudBackupDeleteAllResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupDeleteRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupDeleteRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupDeleteRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupDeleteRequest(buffer_arg) {
  return api_pb.SdkCloudBackupDeleteRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupDeleteResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupDeleteResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupDeleteResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupDeleteResponse(buffer_arg) {
  return api_pb.SdkCloudBackupDeleteResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupEnumerateRequest(buffer_arg) {
  return api_pb.SdkCloudBackupEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupEnumerateResponse(buffer_arg) {
  return api_pb.SdkCloudBackupEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupHistoryRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupHistoryRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupHistoryRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupHistoryRequest(buffer_arg) {
  return api_pb.SdkCloudBackupHistoryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupHistoryResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupHistoryResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupHistoryResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupHistoryResponse(buffer_arg) {
  return api_pb.SdkCloudBackupHistoryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupRestoreRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupRestoreRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupRestoreRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupRestoreRequest(buffer_arg) {
  return api_pb.SdkCloudBackupRestoreRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupRestoreResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupRestoreResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupRestoreResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupRestoreResponse(buffer_arg) {
  return api_pb.SdkCloudBackupRestoreResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupStateChangeRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupStateChangeRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupStateChangeRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupStateChangeRequest(buffer_arg) {
  return api_pb.SdkCloudBackupStateChangeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupStateChangeResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupStateChangeResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupStateChangeResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupStateChangeResponse(buffer_arg) {
  return api_pb.SdkCloudBackupStateChangeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupStatusRequest(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupStatusRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupStatusRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupStatusRequest(buffer_arg) {
  return api_pb.SdkCloudBackupStatusRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkCloudBackupStatusResponse(arg) {
  if (!(arg instanceof api_pb.SdkCloudBackupStatusResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkCloudBackupStatusResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkCloudBackupStatusResponse(buffer_arg) {
  return api_pb.SdkCloudBackupStatusResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

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

function serialize_openstorage_api_SdkClusterAlertDeleteRequest(arg) {
  if (!(arg instanceof api_pb.SdkClusterAlertDeleteRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterAlertDeleteRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterAlertDeleteRequest(buffer_arg) {
  return api_pb.SdkClusterAlertDeleteRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkClusterAlertDeleteResponse(arg) {
  if (!(arg instanceof api_pb.SdkClusterAlertDeleteResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkClusterAlertDeleteResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkClusterAlertDeleteResponse(buffer_arg) {
  return api_pb.SdkClusterAlertDeleteResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_openstorage_api_SdkObjectstoreCreateRequest(arg) {
  if (!(arg instanceof api_pb.SdkObjectstoreCreateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkObjectstoreCreateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkObjectstoreCreateRequest(buffer_arg) {
  return api_pb.SdkObjectstoreCreateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkObjectstoreCreateResponse(arg) {
  if (!(arg instanceof api_pb.SdkObjectstoreCreateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkObjectstoreCreateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkObjectstoreCreateResponse(buffer_arg) {
  return api_pb.SdkObjectstoreCreateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkObjectstoreDeleteRequest(arg) {
  if (!(arg instanceof api_pb.SdkObjectstoreDeleteRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkObjectstoreDeleteRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkObjectstoreDeleteRequest(buffer_arg) {
  return api_pb.SdkObjectstoreDeleteRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkObjectstoreDeleteResponse(arg) {
  if (!(arg instanceof api_pb.SdkObjectstoreDeleteResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkObjectstoreDeleteResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkObjectstoreDeleteResponse(buffer_arg) {
  return api_pb.SdkObjectstoreDeleteResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkObjectstoreInspectRequest(arg) {
  if (!(arg instanceof api_pb.SdkObjectstoreInspectRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkObjectstoreInspectRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkObjectstoreInspectRequest(buffer_arg) {
  return api_pb.SdkObjectstoreInspectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkObjectstoreInspectResponse(arg) {
  if (!(arg instanceof api_pb.SdkObjectstoreInspectResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkObjectstoreInspectResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkObjectstoreInspectResponse(buffer_arg) {
  return api_pb.SdkObjectstoreInspectResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkObjectstoreUpdateRequest(arg) {
  if (!(arg instanceof api_pb.SdkObjectstoreUpdateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkObjectstoreUpdateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkObjectstoreUpdateRequest(buffer_arg) {
  return api_pb.SdkObjectstoreUpdateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkObjectstoreUpdateResponse(arg) {
  if (!(arg instanceof api_pb.SdkObjectstoreUpdateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkObjectstoreUpdateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkObjectstoreUpdateResponse(buffer_arg) {
  return api_pb.SdkObjectstoreUpdateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyCreateRequest(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyCreateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyCreateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyCreateRequest(buffer_arg) {
  return api_pb.SdkSchedulePolicyCreateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyCreateResponse(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyCreateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyCreateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyCreateResponse(buffer_arg) {
  return api_pb.SdkSchedulePolicyCreateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyDeleteRequest(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyDeleteRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyDeleteRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyDeleteRequest(buffer_arg) {
  return api_pb.SdkSchedulePolicyDeleteRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyDeleteResponse(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyDeleteResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyDeleteResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyDeleteResponse(buffer_arg) {
  return api_pb.SdkSchedulePolicyDeleteResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyEnumerateRequest(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyEnumerateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyEnumerateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyEnumerateRequest(buffer_arg) {
  return api_pb.SdkSchedulePolicyEnumerateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyEnumerateResponse(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyEnumerateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyEnumerateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyEnumerateResponse(buffer_arg) {
  return api_pb.SdkSchedulePolicyEnumerateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyInspectRequest(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyInspectRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyInspectRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyInspectRequest(buffer_arg) {
  return api_pb.SdkSchedulePolicyInspectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyInspectResponse(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyInspectResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyInspectResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyInspectResponse(buffer_arg) {
  return api_pb.SdkSchedulePolicyInspectResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyUpdateRequest(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyUpdateRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyUpdateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyUpdateRequest(buffer_arg) {
  return api_pb.SdkSchedulePolicyUpdateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkSchedulePolicyUpdateResponse(arg) {
  if (!(arg instanceof api_pb.SdkSchedulePolicyUpdateResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkSchedulePolicyUpdateResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkSchedulePolicyUpdateResponse(buffer_arg) {
  return api_pb.SdkSchedulePolicyUpdateResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_openstorage_api_SdkVolumeCloneRequest(arg) {
  if (!(arg instanceof api_pb.SdkVolumeCloneRequest)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeCloneRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeCloneRequest(buffer_arg) {
  return api_pb.SdkVolumeCloneRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_openstorage_api_SdkVolumeCloneResponse(arg) {
  if (!(arg instanceof api_pb.SdkVolumeCloneResponse)) {
    throw new Error('Expected argument of type openstorage.api.SdkVolumeCloneResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_openstorage_api_SdkVolumeCloneResponse(buffer_arg) {
  return api_pb.SdkVolumeCloneResponse.deserializeBinary(new Uint8Array(buffer_arg));
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
  alertDelete: {
    path: '/openstorage.api.OpenStorageCluster/AlertDelete',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkClusterAlertDeleteRequest,
    responseType: api_pb.SdkClusterAlertDeleteResponse,
    requestSerialize: serialize_openstorage_api_SdkClusterAlertDeleteRequest,
    requestDeserialize: deserialize_openstorage_api_SdkClusterAlertDeleteRequest,
    responseSerialize: serialize_openstorage_api_SdkClusterAlertDeleteResponse,
    responseDeserialize: deserialize_openstorage_api_SdkClusterAlertDeleteResponse,
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
  // Clone creates a new volume cloned from an existing volume
  clone: {
    path: '/openstorage.api.OpenStorageVolume/Clone',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkVolumeCloneRequest,
    responseType: api_pb.SdkVolumeCloneResponse,
    requestSerialize: serialize_openstorage_api_SdkVolumeCloneRequest,
    requestDeserialize: deserialize_openstorage_api_SdkVolumeCloneRequest,
    responseSerialize: serialize_openstorage_api_SdkVolumeCloneResponse,
    responseDeserialize: deserialize_openstorage_api_SdkVolumeCloneResponse,
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
var OpenStorageObjectstoreService = exports.OpenStorageObjectstoreService = {
  // Inspect returns current status of objectstore
  inspect: {
    path: '/openstorage.api.OpenStorageObjectstore/Inspect',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkObjectstoreInspectRequest,
    responseType: api_pb.SdkObjectstoreInspectResponse,
    requestSerialize: serialize_openstorage_api_SdkObjectstoreInspectRequest,
    requestDeserialize: deserialize_openstorage_api_SdkObjectstoreInspectRequest,
    responseSerialize: serialize_openstorage_api_SdkObjectstoreInspectResponse,
    responseDeserialize: deserialize_openstorage_api_SdkObjectstoreInspectResponse,
  },
  // Creates objectstore on specified volume
  create: {
    path: '/openstorage.api.OpenStorageObjectstore/Create',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkObjectstoreCreateRequest,
    responseType: api_pb.SdkObjectstoreCreateResponse,
    requestSerialize: serialize_openstorage_api_SdkObjectstoreCreateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkObjectstoreCreateRequest,
    responseSerialize: serialize_openstorage_api_SdkObjectstoreCreateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkObjectstoreCreateResponse,
  },
  // Deletes objectstore by id
  delete: {
    path: '/openstorage.api.OpenStorageObjectstore/Delete',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkObjectstoreDeleteRequest,
    responseType: api_pb.SdkObjectstoreDeleteResponse,
    requestSerialize: serialize_openstorage_api_SdkObjectstoreDeleteRequest,
    requestDeserialize: deserialize_openstorage_api_SdkObjectstoreDeleteRequest,
    responseSerialize: serialize_openstorage_api_SdkObjectstoreDeleteResponse,
    responseDeserialize: deserialize_openstorage_api_SdkObjectstoreDeleteResponse,
  },
  // Updates provided objectstore status
  update: {
    path: '/openstorage.api.OpenStorageObjectstore/Update',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkObjectstoreUpdateRequest,
    responseType: api_pb.SdkObjectstoreUpdateResponse,
    requestSerialize: serialize_openstorage_api_SdkObjectstoreUpdateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkObjectstoreUpdateRequest,
    responseSerialize: serialize_openstorage_api_SdkObjectstoreUpdateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkObjectstoreUpdateResponse,
  },
};

exports.OpenStorageObjectstoreClient = grpc.makeGenericClientConstructor(OpenStorageObjectstoreService);
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
  delete: {
    path: '/openstorage.api.OpenStorageCredentials/Delete',
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
  validate: {
    path: '/openstorage.api.OpenStorageCredentials/Validate',
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
var OpenStorageSchedulePolicyService = exports.OpenStorageSchedulePolicyService = {
  // Create Schedule Policy for snapshots
  create: {
    path: '/openstorage.api.OpenStorageSchedulePolicy/Create',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkSchedulePolicyCreateRequest,
    responseType: api_pb.SdkSchedulePolicyCreateResponse,
    requestSerialize: serialize_openstorage_api_SdkSchedulePolicyCreateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkSchedulePolicyCreateRequest,
    responseSerialize: serialize_openstorage_api_SdkSchedulePolicyCreateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkSchedulePolicyCreateResponse,
  },
  // Update Schedule Policy
  update: {
    path: '/openstorage.api.OpenStorageSchedulePolicy/Update',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkSchedulePolicyUpdateRequest,
    responseType: api_pb.SdkSchedulePolicyUpdateResponse,
    requestSerialize: serialize_openstorage_api_SdkSchedulePolicyUpdateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkSchedulePolicyUpdateRequest,
    responseSerialize: serialize_openstorage_api_SdkSchedulePolicyUpdateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkSchedulePolicyUpdateResponse,
  },
  enumerate: {
    path: '/openstorage.api.OpenStorageSchedulePolicy/Enumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkSchedulePolicyEnumerateRequest,
    responseType: api_pb.SdkSchedulePolicyEnumerateResponse,
    requestSerialize: serialize_openstorage_api_SdkSchedulePolicyEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkSchedulePolicyEnumerateRequest,
    responseSerialize: serialize_openstorage_api_SdkSchedulePolicyEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkSchedulePolicyEnumerateResponse,
  },
  // Inspect Schedule Policy
  inspect: {
    path: '/openstorage.api.OpenStorageSchedulePolicy/Inspect',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkSchedulePolicyInspectRequest,
    responseType: api_pb.SdkSchedulePolicyInspectResponse,
    requestSerialize: serialize_openstorage_api_SdkSchedulePolicyInspectRequest,
    requestDeserialize: deserialize_openstorage_api_SdkSchedulePolicyInspectRequest,
    responseSerialize: serialize_openstorage_api_SdkSchedulePolicyInspectResponse,
    responseDeserialize: deserialize_openstorage_api_SdkSchedulePolicyInspectResponse,
  },
  // Delete Schedule Policy
  delete: {
    path: '/openstorage.api.OpenStorageSchedulePolicy/Delete',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkSchedulePolicyDeleteRequest,
    responseType: api_pb.SdkSchedulePolicyDeleteResponse,
    requestSerialize: serialize_openstorage_api_SdkSchedulePolicyDeleteRequest,
    requestDeserialize: deserialize_openstorage_api_SdkSchedulePolicyDeleteRequest,
    responseSerialize: serialize_openstorage_api_SdkSchedulePolicyDeleteResponse,
    responseDeserialize: deserialize_openstorage_api_SdkSchedulePolicyDeleteResponse,
  },
};

exports.OpenStorageSchedulePolicyClient = grpc.makeGenericClientConstructor(OpenStorageSchedulePolicyService);
var OpenStorageCloudBackupService = exports.OpenStorageCloudBackupService = {
  // Create
  create: {
    path: '/openstorage.api.OpenStorageCloudBackup/Create',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupCreateRequest,
    responseType: api_pb.SdkCloudBackupCreateResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupCreateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupCreateRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupCreateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupCreateResponse,
  },
  // Restore
  restore: {
    path: '/openstorage.api.OpenStorageCloudBackup/Restore',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupRestoreRequest,
    responseType: api_pb.SdkCloudBackupRestoreResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupRestoreRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupRestoreRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupRestoreResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupRestoreResponse,
  },
  // Delete
  delete: {
    path: '/openstorage.api.OpenStorageCloudBackup/Delete',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupDeleteRequest,
    responseType: api_pb.SdkCloudBackupDeleteResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupDeleteRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupDeleteRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupDeleteResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupDeleteResponse,
  },
  // DeleteAll
  deleteAll: {
    path: '/openstorage.api.OpenStorageCloudBackup/DeleteAll',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupDeleteAllRequest,
    responseType: api_pb.SdkCloudBackupDeleteAllResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupDeleteAllRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupDeleteAllRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupDeleteAllResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupDeleteAllResponse,
  },
  // Enumerate
  enumerate: {
    path: '/openstorage.api.OpenStorageCloudBackup/Enumerate',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupEnumerateRequest,
    responseType: api_pb.SdkCloudBackupEnumerateResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupEnumerateRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupEnumerateRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupEnumerateResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupEnumerateResponse,
  },
  // Status
  status: {
    path: '/openstorage.api.OpenStorageCloudBackup/Status',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupStatusRequest,
    responseType: api_pb.SdkCloudBackupStatusResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupStatusRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupStatusRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupStatusResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupStatusResponse,
  },
  // Catalog
  catalog: {
    path: '/openstorage.api.OpenStorageCloudBackup/Catalog',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupCatalogRequest,
    responseType: api_pb.SdkCloudBackupCatalogResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupCatalogRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupCatalogRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupCatalogResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupCatalogResponse,
  },
  // History
  history: {
    path: '/openstorage.api.OpenStorageCloudBackup/History',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupHistoryRequest,
    responseType: api_pb.SdkCloudBackupHistoryResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupHistoryRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupHistoryRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupHistoryResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupHistoryResponse,
  },
  // StateChange
  stateChange: {
    path: '/openstorage.api.OpenStorageCloudBackup/StateChange',
    requestStream: false,
    responseStream: false,
    requestType: api_pb.SdkCloudBackupStateChangeRequest,
    responseType: api_pb.SdkCloudBackupStateChangeResponse,
    requestSerialize: serialize_openstorage_api_SdkCloudBackupStateChangeRequest,
    requestDeserialize: deserialize_openstorage_api_SdkCloudBackupStateChangeRequest,
    responseSerialize: serialize_openstorage_api_SdkCloudBackupStateChangeResponse,
    responseDeserialize: deserialize_openstorage_api_SdkCloudBackupStateChangeResponse,
  },
};

exports.OpenStorageCloudBackupClient = grpc.makeGenericClientConstructor(OpenStorageCloudBackupService);
