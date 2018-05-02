/**
 * @fileoverview
 * @enhanceable
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!

var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
goog.exportSymbol('proto.openstorage.api.ActiveRequest', null, global);
goog.exportSymbol('proto.openstorage.api.ActiveRequests', null, global);
goog.exportSymbol('proto.openstorage.api.Alert', null, global);
goog.exportSymbol('proto.openstorage.api.AlertActionType', null, global);
goog.exportSymbol('proto.openstorage.api.Alerts', null, global);
goog.exportSymbol('proto.openstorage.api.AttachState', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterAlertClearRequest', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterAlertClearResponse', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterAlertEnumerateRequest', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterAlertEnumerateResponse', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterAlertEraseRequest', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterAlertEraseResponse', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterEnumerateRequest', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterEnumerateResponse', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterInspectRequest', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterInspectResponse', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterNotify', null, global);
goog.exportSymbol('proto.openstorage.api.ClusterResponse', null, global);
goog.exportSymbol('proto.openstorage.api.CosType', null, global);
goog.exportSymbol('proto.openstorage.api.DriverType', null, global);
goog.exportSymbol('proto.openstorage.api.FSType', null, global);
goog.exportSymbol('proto.openstorage.api.GraphDriverChangeType', null, global);
goog.exportSymbol('proto.openstorage.api.GraphDriverChanges', null, global);
goog.exportSymbol('proto.openstorage.api.Group', null, global);
goog.exportSymbol('proto.openstorage.api.GroupSnapCreateRequest', null, global);
goog.exportSymbol('proto.openstorage.api.GroupSnapCreateResponse', null, global);
goog.exportSymbol('proto.openstorage.api.IoProfile', null, global);
goog.exportSymbol('proto.openstorage.api.OpenStorageVolumeCreateRequest', null, global);
goog.exportSymbol('proto.openstorage.api.OpenStorageVolumeCreateResponse', null, global);
goog.exportSymbol('proto.openstorage.api.OperationFlags', null, global);
goog.exportSymbol('proto.openstorage.api.ReplicaSet', null, global);
goog.exportSymbol('proto.openstorage.api.ResourceType', null, global);
goog.exportSymbol('proto.openstorage.api.RuntimeStateMap', null, global);
goog.exportSymbol('proto.openstorage.api.SeverityType', null, global);
goog.exportSymbol('proto.openstorage.api.SnapCreateRequest', null, global);
goog.exportSymbol('proto.openstorage.api.SnapCreateResponse', null, global);
goog.exportSymbol('proto.openstorage.api.Source', null, global);
goog.exportSymbol('proto.openstorage.api.Stats', null, global);
goog.exportSymbol('proto.openstorage.api.Status', null, global);
goog.exportSymbol('proto.openstorage.api.StorageCluster', null, global);
goog.exportSymbol('proto.openstorage.api.StorageMedium', null, global);
goog.exportSymbol('proto.openstorage.api.StorageNode', null, global);
goog.exportSymbol('proto.openstorage.api.StoragePool', null, global);
goog.exportSymbol('proto.openstorage.api.StorageResource', null, global);
goog.exportSymbol('proto.openstorage.api.Volume', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeActionParam', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeConsumer', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeCreateFromVolumeIDRequest', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeCreateFromVolumeIDResponse', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeCreateRequest', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeCreateResponse', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeDeleteRequest', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeDeleteResponse', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeEnumerateRequest', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeEnumerateResponse', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeInfo', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeInspectRequest', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeInspectResponse', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeLocator', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeResponse', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeSetRequest', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeSetResponse', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeSpec', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeState', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeStateAction', null, global);
goog.exportSymbol('proto.openstorage.api.VolumeStatus', null, global);

/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.StorageResource = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.StorageResource, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.StorageResource.displayName = 'proto.openstorage.api.StorageResource';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.StorageResource.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.StorageResource.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.StorageResource} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.StorageResource.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    path: jspb.Message.getFieldWithDefault(msg, 2, ""),
    medium: jspb.Message.getFieldWithDefault(msg, 3, 0),
    online: jspb.Message.getFieldWithDefault(msg, 4, false),
    iops: jspb.Message.getFieldWithDefault(msg, 5, 0),
    seqWrite: +jspb.Message.getFieldWithDefault(msg, 6, 0.0),
    seqRead: +jspb.Message.getFieldWithDefault(msg, 7, 0.0),
    randrw: +jspb.Message.getFieldWithDefault(msg, 8, 0.0),
    size: jspb.Message.getFieldWithDefault(msg, 9, 0),
    used: jspb.Message.getFieldWithDefault(msg, 10, 0),
    rotationSpeed: jspb.Message.getFieldWithDefault(msg, 11, ""),
    lastScan: (f = msg.getLastScan()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    metadata: jspb.Message.getFieldWithDefault(msg, 13, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.StorageResource}
 */
proto.openstorage.api.StorageResource.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.StorageResource;
  return proto.openstorage.api.StorageResource.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.StorageResource} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.StorageResource}
 */
proto.openstorage.api.StorageResource.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
      break;
    case 3:
      var value = /** @type {!proto.openstorage.api.StorageMedium} */ (reader.readEnum());
      msg.setMedium(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setOnline(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setIops(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setSeqWrite(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setSeqRead(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setRandrw(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSize(value);
      break;
    case 10:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setUsed(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setRotationSpeed(value);
      break;
    case 12:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setLastScan(value);
      break;
    case 13:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setMetadata(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.StorageResource.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.StorageResource.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.StorageResource} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.StorageResource.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getMedium();
  if (f !== 0.0) {
    writer.writeEnum(
      3,
      f
    );
  }
  f = message.getOnline();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
  f = message.getIops();
  if (f !== 0) {
    writer.writeUint64(
      5,
      f
    );
  }
  f = message.getSeqWrite();
  if (f !== 0.0) {
    writer.writeDouble(
      6,
      f
    );
  }
  f = message.getSeqRead();
  if (f !== 0.0) {
    writer.writeDouble(
      7,
      f
    );
  }
  f = message.getRandrw();
  if (f !== 0.0) {
    writer.writeDouble(
      8,
      f
    );
  }
  f = message.getSize();
  if (f !== 0) {
    writer.writeUint64(
      9,
      f
    );
  }
  f = message.getUsed();
  if (f !== 0) {
    writer.writeUint64(
      10,
      f
    );
  }
  f = message.getRotationSpeed();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
  f = message.getLastScan();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getMetadata();
  if (f) {
    writer.writeBool(
      13,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.openstorage.api.StorageResource.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageResource.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional string path = 2;
 * @return {string}
 */
proto.openstorage.api.StorageResource.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageResource.prototype.setPath = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional StorageMedium medium = 3;
 * @return {!proto.openstorage.api.StorageMedium}
 */
proto.openstorage.api.StorageResource.prototype.getMedium = function() {
  return /** @type {!proto.openstorage.api.StorageMedium} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/** @param {!proto.openstorage.api.StorageMedium} value */
proto.openstorage.api.StorageResource.prototype.setMedium = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * optional bool online = 4;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.StorageResource.prototype.getOnline = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 4, false));
};


/** @param {boolean} value */
proto.openstorage.api.StorageResource.prototype.setOnline = function(value) {
  jspb.Message.setField(this, 4, value);
};


/**
 * optional uint64 iops = 5;
 * @return {number}
 */
proto.openstorage.api.StorageResource.prototype.getIops = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/** @param {number} value */
proto.openstorage.api.StorageResource.prototype.setIops = function(value) {
  jspb.Message.setField(this, 5, value);
};


/**
 * optional double seq_write = 6;
 * @return {number}
 */
proto.openstorage.api.StorageResource.prototype.getSeqWrite = function() {
  return /** @type {number} */ (+jspb.Message.getFieldWithDefault(this, 6, 0.0));
};


/** @param {number} value */
proto.openstorage.api.StorageResource.prototype.setSeqWrite = function(value) {
  jspb.Message.setField(this, 6, value);
};


/**
 * optional double seq_read = 7;
 * @return {number}
 */
proto.openstorage.api.StorageResource.prototype.getSeqRead = function() {
  return /** @type {number} */ (+jspb.Message.getFieldWithDefault(this, 7, 0.0));
};


/** @param {number} value */
proto.openstorage.api.StorageResource.prototype.setSeqRead = function(value) {
  jspb.Message.setField(this, 7, value);
};


/**
 * optional double randRW = 8;
 * @return {number}
 */
proto.openstorage.api.StorageResource.prototype.getRandrw = function() {
  return /** @type {number} */ (+jspb.Message.getFieldWithDefault(this, 8, 0.0));
};


/** @param {number} value */
proto.openstorage.api.StorageResource.prototype.setRandrw = function(value) {
  jspb.Message.setField(this, 8, value);
};


/**
 * optional uint64 size = 9;
 * @return {number}
 */
proto.openstorage.api.StorageResource.prototype.getSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/** @param {number} value */
proto.openstorage.api.StorageResource.prototype.setSize = function(value) {
  jspb.Message.setField(this, 9, value);
};


/**
 * optional uint64 used = 10;
 * @return {number}
 */
proto.openstorage.api.StorageResource.prototype.getUsed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 10, 0));
};


/** @param {number} value */
proto.openstorage.api.StorageResource.prototype.setUsed = function(value) {
  jspb.Message.setField(this, 10, value);
};


/**
 * optional string rotation_speed = 11;
 * @return {string}
 */
proto.openstorage.api.StorageResource.prototype.getRotationSpeed = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageResource.prototype.setRotationSpeed = function(value) {
  jspb.Message.setField(this, 11, value);
};


/**
 * optional google.protobuf.Timestamp last_scan = 12;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.openstorage.api.StorageResource.prototype.getLastScan = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 12));
};


/** @param {?proto.google.protobuf.Timestamp|undefined} value */
proto.openstorage.api.StorageResource.prototype.setLastScan = function(value) {
  jspb.Message.setWrapperField(this, 12, value);
};


proto.openstorage.api.StorageResource.prototype.clearLastScan = function() {
  this.setLastScan(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.StorageResource.prototype.hasLastScan = function() {
  return jspb.Message.getField(this, 12) != null;
};


/**
 * optional bool metadata = 13;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.StorageResource.prototype.getMetadata = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 13, false));
};


/** @param {boolean} value */
proto.openstorage.api.StorageResource.prototype.setMetadata = function(value) {
  jspb.Message.setField(this, 13, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.StoragePool = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.StoragePool, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.StoragePool.displayName = 'proto.openstorage.api.StoragePool';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.StoragePool.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.StoragePool.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.StoragePool} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.StoragePool.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, 0),
    cos: jspb.Message.getFieldWithDefault(msg, 2, 0),
    medium: jspb.Message.getFieldWithDefault(msg, 3, 0),
    raidlevel: jspb.Message.getFieldWithDefault(msg, 4, ""),
    totalsize: jspb.Message.getFieldWithDefault(msg, 7, 0),
    used: jspb.Message.getFieldWithDefault(msg, 8, 0),
    labelsMap: (f = msg.getLabelsMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.StoragePool}
 */
proto.openstorage.api.StoragePool.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.StoragePool;
  return proto.openstorage.api.StoragePool.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.StoragePool} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.StoragePool}
 */
proto.openstorage.api.StoragePool.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {!proto.openstorage.api.CosType} */ (reader.readEnum());
      msg.setCos(value);
      break;
    case 3:
      var value = /** @type {!proto.openstorage.api.StorageMedium} */ (reader.readEnum());
      msg.setMedium(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setRaidlevel(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setTotalsize(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setUsed(value);
      break;
    case 9:
      var value = msg.getLabelsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.StoragePool.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.StoragePool.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.StoragePool} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.StoragePool.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f !== 0) {
    writer.writeInt32(
      1,
      f
    );
  }
  f = message.getCos();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getMedium();
  if (f !== 0.0) {
    writer.writeEnum(
      3,
      f
    );
  }
  f = message.getRaidlevel();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getTotalsize();
  if (f !== 0) {
    writer.writeUint64(
      7,
      f
    );
  }
  f = message.getUsed();
  if (f !== 0) {
    writer.writeUint64(
      8,
      f
    );
  }
  f = message.getLabelsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(9, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * optional int32 ID = 1;
 * @return {number}
 */
proto.openstorage.api.StoragePool.prototype.getId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {number} value */
proto.openstorage.api.StoragePool.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional CosType Cos = 2;
 * @return {!proto.openstorage.api.CosType}
 */
proto.openstorage.api.StoragePool.prototype.getCos = function() {
  return /** @type {!proto.openstorage.api.CosType} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/** @param {!proto.openstorage.api.CosType} value */
proto.openstorage.api.StoragePool.prototype.setCos = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional StorageMedium Medium = 3;
 * @return {!proto.openstorage.api.StorageMedium}
 */
proto.openstorage.api.StoragePool.prototype.getMedium = function() {
  return /** @type {!proto.openstorage.api.StorageMedium} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/** @param {!proto.openstorage.api.StorageMedium} value */
proto.openstorage.api.StoragePool.prototype.setMedium = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * optional string RaidLevel = 4;
 * @return {string}
 */
proto.openstorage.api.StoragePool.prototype.getRaidlevel = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/** @param {string} value */
proto.openstorage.api.StoragePool.prototype.setRaidlevel = function(value) {
  jspb.Message.setField(this, 4, value);
};


/**
 * optional uint64 TotalSize = 7;
 * @return {number}
 */
proto.openstorage.api.StoragePool.prototype.getTotalsize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/** @param {number} value */
proto.openstorage.api.StoragePool.prototype.setTotalsize = function(value) {
  jspb.Message.setField(this, 7, value);
};


/**
 * optional uint64 Used = 8;
 * @return {number}
 */
proto.openstorage.api.StoragePool.prototype.getUsed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/** @param {number} value */
proto.openstorage.api.StoragePool.prototype.setUsed = function(value) {
  jspb.Message.setField(this, 8, value);
};


/**
 * map<string, string> labels = 9;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.openstorage.api.StoragePool.prototype.getLabelsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 9, opt_noLazyCreate,
      null));
};


proto.openstorage.api.StoragePool.prototype.clearLabelsMap = function() {
  this.getLabelsMap().clear();
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeLocator = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeLocator, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeLocator.displayName = 'proto.openstorage.api.VolumeLocator';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeLocator.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeLocator.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeLocator} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeLocator.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    volumeLabelsMap: (f = msg.getVolumeLabelsMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeLocator}
 */
proto.openstorage.api.VolumeLocator.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeLocator;
  return proto.openstorage.api.VolumeLocator.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeLocator} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeLocator}
 */
proto.openstorage.api.VolumeLocator.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 2:
      var value = msg.getVolumeLabelsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeLocator.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeLocator.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeLocator} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeLocator.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getVolumeLabelsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(2, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeLocator.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeLocator.prototype.setName = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * map<string, string> volume_labels = 2;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.openstorage.api.VolumeLocator.prototype.getVolumeLabelsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 2, opt_noLazyCreate,
      null));
};


proto.openstorage.api.VolumeLocator.prototype.clearVolumeLabelsMap = function() {
  this.getVolumeLabelsMap().clear();
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.Source = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.Source, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.Source.displayName = 'proto.openstorage.api.Source';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.Source.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.Source.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.Source} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Source.toObject = function(includeInstance, msg) {
  var f, obj = {
    parent: jspb.Message.getFieldWithDefault(msg, 1, ""),
    seed: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.Source}
 */
proto.openstorage.api.Source.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.Source;
  return proto.openstorage.api.Source.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.Source} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.Source}
 */
proto.openstorage.api.Source.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setParent(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setSeed(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.Source.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.Source.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.Source} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Source.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getParent();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSeed();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional string parent = 1;
 * @return {string}
 */
proto.openstorage.api.Source.prototype.getParent = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.Source.prototype.setParent = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional string seed = 2;
 * @return {string}
 */
proto.openstorage.api.Source.prototype.getSeed = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/** @param {string} value */
proto.openstorage.api.Source.prototype.setSeed = function(value) {
  jspb.Message.setField(this, 2, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.Group = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.Group, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.Group.displayName = 'proto.openstorage.api.Group';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.Group.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.Group.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.Group} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Group.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.Group}
 */
proto.openstorage.api.Group.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.Group;
  return proto.openstorage.api.Group.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.Group} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.Group}
 */
proto.openstorage.api.Group.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.Group.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.Group.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.Group} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Group.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.openstorage.api.Group.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.Group.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeSpec = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeSpec, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeSpec.displayName = 'proto.openstorage.api.VolumeSpec';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeSpec.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeSpec.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeSpec} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeSpec.toObject = function(includeInstance, msg) {
  var f, obj = {
    ephemeral: jspb.Message.getFieldWithDefault(msg, 1, false),
    size: jspb.Message.getFieldWithDefault(msg, 2, 0),
    format: jspb.Message.getFieldWithDefault(msg, 3, 0),
    blockSize: jspb.Message.getFieldWithDefault(msg, 4, 0),
    haLevel: jspb.Message.getFieldWithDefault(msg, 5, 0),
    cos: jspb.Message.getFieldWithDefault(msg, 6, 0),
    ioProfile: jspb.Message.getFieldWithDefault(msg, 7, 0),
    dedupe: jspb.Message.getFieldWithDefault(msg, 8, false),
    snapshotInterval: jspb.Message.getFieldWithDefault(msg, 9, 0),
    volumeLabelsMap: (f = msg.getVolumeLabelsMap()) ? f.toObject(includeInstance, undefined) : [],
    shared: jspb.Message.getFieldWithDefault(msg, 11, false),
    replicaSet: (f = msg.getReplicaSet()) && proto.openstorage.api.ReplicaSet.toObject(includeInstance, f),
    aggregationLevel: jspb.Message.getFieldWithDefault(msg, 13, 0),
    encrypted: jspb.Message.getFieldWithDefault(msg, 14, false),
    passphrase: jspb.Message.getFieldWithDefault(msg, 15, ""),
    snapshotSchedule: jspb.Message.getFieldWithDefault(msg, 16, ""),
    scale: jspb.Message.getFieldWithDefault(msg, 17, 0),
    sticky: jspb.Message.getFieldWithDefault(msg, 18, false),
    group: (f = msg.getGroup()) && proto.openstorage.api.Group.toObject(includeInstance, f),
    groupEnforced: jspb.Message.getFieldWithDefault(msg, 22, false),
    compressed: jspb.Message.getFieldWithDefault(msg, 23, false),
    cascaded: jspb.Message.getFieldWithDefault(msg, 24, false),
    journal: jspb.Message.getFieldWithDefault(msg, 25, false),
    sharedv4: jspb.Message.getFieldWithDefault(msg, 26, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeSpec}
 */
proto.openstorage.api.VolumeSpec.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeSpec;
  return proto.openstorage.api.VolumeSpec.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeSpec} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeSpec}
 */
proto.openstorage.api.VolumeSpec.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEphemeral(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setSize(value);
      break;
    case 3:
      var value = /** @type {!proto.openstorage.api.FSType} */ (reader.readEnum());
      msg.setFormat(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setBlockSize(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setHaLevel(value);
      break;
    case 6:
      var value = /** @type {!proto.openstorage.api.CosType} */ (reader.readEnum());
      msg.setCos(value);
      break;
    case 7:
      var value = /** @type {!proto.openstorage.api.IoProfile} */ (reader.readEnum());
      msg.setIoProfile(value);
      break;
    case 8:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setDedupe(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setSnapshotInterval(value);
      break;
    case 10:
      var value = msg.getVolumeLabelsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString);
         });
      break;
    case 11:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setShared(value);
      break;
    case 12:
      var value = new proto.openstorage.api.ReplicaSet;
      reader.readMessage(value,proto.openstorage.api.ReplicaSet.deserializeBinaryFromReader);
      msg.setReplicaSet(value);
      break;
    case 13:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setAggregationLevel(value);
      break;
    case 14:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setEncrypted(value);
      break;
    case 15:
      var value = /** @type {string} */ (reader.readString());
      msg.setPassphrase(value);
      break;
    case 16:
      var value = /** @type {string} */ (reader.readString());
      msg.setSnapshotSchedule(value);
      break;
    case 17:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setScale(value);
      break;
    case 18:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSticky(value);
      break;
    case 21:
      var value = new proto.openstorage.api.Group;
      reader.readMessage(value,proto.openstorage.api.Group.deserializeBinaryFromReader);
      msg.setGroup(value);
      break;
    case 22:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setGroupEnforced(value);
      break;
    case 23:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setCompressed(value);
      break;
    case 24:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setCascaded(value);
      break;
    case 25:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setJournal(value);
      break;
    case 26:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setSharedv4(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeSpec.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeSpec.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeSpec} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeSpec.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getEphemeral();
  if (f) {
    writer.writeBool(
      1,
      f
    );
  }
  f = message.getSize();
  if (f !== 0) {
    writer.writeUint64(
      2,
      f
    );
  }
  f = message.getFormat();
  if (f !== 0.0) {
    writer.writeEnum(
      3,
      f
    );
  }
  f = message.getBlockSize();
  if (f !== 0) {
    writer.writeInt64(
      4,
      f
    );
  }
  f = message.getHaLevel();
  if (f !== 0) {
    writer.writeInt64(
      5,
      f
    );
  }
  f = message.getCos();
  if (f !== 0.0) {
    writer.writeEnum(
      6,
      f
    );
  }
  f = message.getIoProfile();
  if (f !== 0.0) {
    writer.writeEnum(
      7,
      f
    );
  }
  f = message.getDedupe();
  if (f) {
    writer.writeBool(
      8,
      f
    );
  }
  f = message.getSnapshotInterval();
  if (f !== 0) {
    writer.writeUint32(
      9,
      f
    );
  }
  f = message.getVolumeLabelsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(10, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getShared();
  if (f) {
    writer.writeBool(
      11,
      f
    );
  }
  f = message.getReplicaSet();
  if (f != null) {
    writer.writeMessage(
      12,
      f,
      proto.openstorage.api.ReplicaSet.serializeBinaryToWriter
    );
  }
  f = message.getAggregationLevel();
  if (f !== 0) {
    writer.writeUint32(
      13,
      f
    );
  }
  f = message.getEncrypted();
  if (f) {
    writer.writeBool(
      14,
      f
    );
  }
  f = message.getPassphrase();
  if (f.length > 0) {
    writer.writeString(
      15,
      f
    );
  }
  f = message.getSnapshotSchedule();
  if (f.length > 0) {
    writer.writeString(
      16,
      f
    );
  }
  f = message.getScale();
  if (f !== 0) {
    writer.writeUint32(
      17,
      f
    );
  }
  f = message.getSticky();
  if (f) {
    writer.writeBool(
      18,
      f
    );
  }
  f = message.getGroup();
  if (f != null) {
    writer.writeMessage(
      21,
      f,
      proto.openstorage.api.Group.serializeBinaryToWriter
    );
  }
  f = message.getGroupEnforced();
  if (f) {
    writer.writeBool(
      22,
      f
    );
  }
  f = message.getCompressed();
  if (f) {
    writer.writeBool(
      23,
      f
    );
  }
  f = message.getCascaded();
  if (f) {
    writer.writeBool(
      24,
      f
    );
  }
  f = message.getJournal();
  if (f) {
    writer.writeBool(
      25,
      f
    );
  }
  f = message.getSharedv4();
  if (f) {
    writer.writeBool(
      26,
      f
    );
  }
};


/**
 * optional bool ephemeral = 1;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getEphemeral = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 1, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setEphemeral = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional uint64 size = 2;
 * @return {number}
 */
proto.openstorage.api.VolumeSpec.prototype.getSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/** @param {number} value */
proto.openstorage.api.VolumeSpec.prototype.setSize = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional FSType format = 3;
 * @return {!proto.openstorage.api.FSType}
 */
proto.openstorage.api.VolumeSpec.prototype.getFormat = function() {
  return /** @type {!proto.openstorage.api.FSType} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/** @param {!proto.openstorage.api.FSType} value */
proto.openstorage.api.VolumeSpec.prototype.setFormat = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * optional int64 block_size = 4;
 * @return {number}
 */
proto.openstorage.api.VolumeSpec.prototype.getBlockSize = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/** @param {number} value */
proto.openstorage.api.VolumeSpec.prototype.setBlockSize = function(value) {
  jspb.Message.setField(this, 4, value);
};


/**
 * optional int64 ha_level = 5;
 * @return {number}
 */
proto.openstorage.api.VolumeSpec.prototype.getHaLevel = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/** @param {number} value */
proto.openstorage.api.VolumeSpec.prototype.setHaLevel = function(value) {
  jspb.Message.setField(this, 5, value);
};


/**
 * optional CosType cos = 6;
 * @return {!proto.openstorage.api.CosType}
 */
proto.openstorage.api.VolumeSpec.prototype.getCos = function() {
  return /** @type {!proto.openstorage.api.CosType} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/** @param {!proto.openstorage.api.CosType} value */
proto.openstorage.api.VolumeSpec.prototype.setCos = function(value) {
  jspb.Message.setField(this, 6, value);
};


/**
 * optional IoProfile io_profile = 7;
 * @return {!proto.openstorage.api.IoProfile}
 */
proto.openstorage.api.VolumeSpec.prototype.getIoProfile = function() {
  return /** @type {!proto.openstorage.api.IoProfile} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/** @param {!proto.openstorage.api.IoProfile} value */
proto.openstorage.api.VolumeSpec.prototype.setIoProfile = function(value) {
  jspb.Message.setField(this, 7, value);
};


/**
 * optional bool dedupe = 8;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getDedupe = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 8, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setDedupe = function(value) {
  jspb.Message.setField(this, 8, value);
};


/**
 * optional uint32 snapshot_interval = 9;
 * @return {number}
 */
proto.openstorage.api.VolumeSpec.prototype.getSnapshotInterval = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/** @param {number} value */
proto.openstorage.api.VolumeSpec.prototype.setSnapshotInterval = function(value) {
  jspb.Message.setField(this, 9, value);
};


/**
 * map<string, string> volume_labels = 10;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.openstorage.api.VolumeSpec.prototype.getVolumeLabelsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 10, opt_noLazyCreate,
      null));
};


proto.openstorage.api.VolumeSpec.prototype.clearVolumeLabelsMap = function() {
  this.getVolumeLabelsMap().clear();
};


/**
 * optional bool shared = 11;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getShared = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 11, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setShared = function(value) {
  jspb.Message.setField(this, 11, value);
};


/**
 * optional ReplicaSet replica_set = 12;
 * @return {?proto.openstorage.api.ReplicaSet}
 */
proto.openstorage.api.VolumeSpec.prototype.getReplicaSet = function() {
  return /** @type{?proto.openstorage.api.ReplicaSet} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.ReplicaSet, 12));
};


/** @param {?proto.openstorage.api.ReplicaSet|undefined} value */
proto.openstorage.api.VolumeSpec.prototype.setReplicaSet = function(value) {
  jspb.Message.setWrapperField(this, 12, value);
};


proto.openstorage.api.VolumeSpec.prototype.clearReplicaSet = function() {
  this.setReplicaSet(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.hasReplicaSet = function() {
  return jspb.Message.getField(this, 12) != null;
};


/**
 * optional uint32 aggregation_level = 13;
 * @return {number}
 */
proto.openstorage.api.VolumeSpec.prototype.getAggregationLevel = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 13, 0));
};


/** @param {number} value */
proto.openstorage.api.VolumeSpec.prototype.setAggregationLevel = function(value) {
  jspb.Message.setField(this, 13, value);
};


/**
 * optional bool encrypted = 14;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getEncrypted = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 14, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setEncrypted = function(value) {
  jspb.Message.setField(this, 14, value);
};


/**
 * optional string passphrase = 15;
 * @return {string}
 */
proto.openstorage.api.VolumeSpec.prototype.getPassphrase = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 15, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeSpec.prototype.setPassphrase = function(value) {
  jspb.Message.setField(this, 15, value);
};


/**
 * optional string snapshot_schedule = 16;
 * @return {string}
 */
proto.openstorage.api.VolumeSpec.prototype.getSnapshotSchedule = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 16, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeSpec.prototype.setSnapshotSchedule = function(value) {
  jspb.Message.setField(this, 16, value);
};


/**
 * optional uint32 scale = 17;
 * @return {number}
 */
proto.openstorage.api.VolumeSpec.prototype.getScale = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 17, 0));
};


/** @param {number} value */
proto.openstorage.api.VolumeSpec.prototype.setScale = function(value) {
  jspb.Message.setField(this, 17, value);
};


/**
 * optional bool sticky = 18;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getSticky = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 18, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setSticky = function(value) {
  jspb.Message.setField(this, 18, value);
};


/**
 * optional Group group = 21;
 * @return {?proto.openstorage.api.Group}
 */
proto.openstorage.api.VolumeSpec.prototype.getGroup = function() {
  return /** @type{?proto.openstorage.api.Group} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.Group, 21));
};


/** @param {?proto.openstorage.api.Group|undefined} value */
proto.openstorage.api.VolumeSpec.prototype.setGroup = function(value) {
  jspb.Message.setWrapperField(this, 21, value);
};


proto.openstorage.api.VolumeSpec.prototype.clearGroup = function() {
  this.setGroup(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.hasGroup = function() {
  return jspb.Message.getField(this, 21) != null;
};


/**
 * optional bool group_enforced = 22;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getGroupEnforced = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 22, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setGroupEnforced = function(value) {
  jspb.Message.setField(this, 22, value);
};


/**
 * optional bool compressed = 23;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getCompressed = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 23, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setCompressed = function(value) {
  jspb.Message.setField(this, 23, value);
};


/**
 * optional bool cascaded = 24;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getCascaded = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 24, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setCascaded = function(value) {
  jspb.Message.setField(this, 24, value);
};


/**
 * optional bool journal = 25;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getJournal = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 25, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setJournal = function(value) {
  jspb.Message.setField(this, 25, value);
};


/**
 * optional bool sharedv4 = 26;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.VolumeSpec.prototype.getSharedv4 = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 26, false));
};


/** @param {boolean} value */
proto.openstorage.api.VolumeSpec.prototype.setSharedv4 = function(value) {
  jspb.Message.setField(this, 26, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ReplicaSet = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.openstorage.api.ReplicaSet.repeatedFields_, null);
};
goog.inherits(proto.openstorage.api.ReplicaSet, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ReplicaSet.displayName = 'proto.openstorage.api.ReplicaSet';
}
/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.openstorage.api.ReplicaSet.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ReplicaSet.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ReplicaSet.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ReplicaSet} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ReplicaSet.toObject = function(includeInstance, msg) {
  var f, obj = {
    nodesList: jspb.Message.getRepeatedField(msg, 1)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ReplicaSet}
 */
proto.openstorage.api.ReplicaSet.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ReplicaSet;
  return proto.openstorage.api.ReplicaSet.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ReplicaSet} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ReplicaSet}
 */
proto.openstorage.api.ReplicaSet.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.addNodes(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ReplicaSet.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ReplicaSet.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ReplicaSet} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ReplicaSet.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNodesList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      1,
      f
    );
  }
};


/**
 * repeated string nodes = 1;
 * @return {!Array.<string>}
 */
proto.openstorage.api.ReplicaSet.prototype.getNodesList = function() {
  return /** @type {!Array.<string>} */ (jspb.Message.getRepeatedField(this, 1));
};


/** @param {!Array.<string>} value */
proto.openstorage.api.ReplicaSet.prototype.setNodesList = function(value) {
  jspb.Message.setField(this, 1, value || []);
};


/**
 * @param {!string} value
 * @param {number=} opt_index
 */
proto.openstorage.api.ReplicaSet.prototype.addNodes = function(value, opt_index) {
  jspb.Message.addToRepeatedField(this, 1, value, opt_index);
};


proto.openstorage.api.ReplicaSet.prototype.clearNodesList = function() {
  this.setNodesList([]);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.RuntimeStateMap = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.RuntimeStateMap, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.RuntimeStateMap.displayName = 'proto.openstorage.api.RuntimeStateMap';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.RuntimeStateMap.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.RuntimeStateMap.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.RuntimeStateMap} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.RuntimeStateMap.toObject = function(includeInstance, msg) {
  var f, obj = {
    runtimeStateMap: (f = msg.getRuntimeStateMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.RuntimeStateMap}
 */
proto.openstorage.api.RuntimeStateMap.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.RuntimeStateMap;
  return proto.openstorage.api.RuntimeStateMap.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.RuntimeStateMap} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.RuntimeStateMap}
 */
proto.openstorage.api.RuntimeStateMap.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getRuntimeStateMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.RuntimeStateMap.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.RuntimeStateMap.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.RuntimeStateMap} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.RuntimeStateMap.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRuntimeStateMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * map<string, string> runtime_state = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.openstorage.api.RuntimeStateMap.prototype.getRuntimeStateMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      null));
};


proto.openstorage.api.RuntimeStateMap.prototype.clearRuntimeStateMap = function() {
  this.getRuntimeStateMap().clear();
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.Volume = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.openstorage.api.Volume.repeatedFields_, null);
};
goog.inherits(proto.openstorage.api.Volume, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.Volume.displayName = 'proto.openstorage.api.Volume';
}
/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.openstorage.api.Volume.repeatedFields_ = [17,19,20,22];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.Volume.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.Volume.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.Volume} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Volume.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    source: (f = msg.getSource()) && proto.openstorage.api.Source.toObject(includeInstance, f),
    group: (f = msg.getGroup()) && proto.openstorage.api.Group.toObject(includeInstance, f),
    readonly: jspb.Message.getFieldWithDefault(msg, 4, false),
    locator: (f = msg.getLocator()) && proto.openstorage.api.VolumeLocator.toObject(includeInstance, f),
    ctime: (f = msg.getCtime()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    spec: (f = msg.getSpec()) && proto.openstorage.api.VolumeSpec.toObject(includeInstance, f),
    usage: jspb.Message.getFieldWithDefault(msg, 8, 0),
    lastScan: (f = msg.getLastScan()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    format: jspb.Message.getFieldWithDefault(msg, 10, 0),
    status: jspb.Message.getFieldWithDefault(msg, 11, 0),
    state: jspb.Message.getFieldWithDefault(msg, 12, 0),
    attachedOn: jspb.Message.getFieldWithDefault(msg, 13, ""),
    attachedState: jspb.Message.getFieldWithDefault(msg, 14, 0),
    devicePath: jspb.Message.getFieldWithDefault(msg, 15, ""),
    secureDevicePath: jspb.Message.getFieldWithDefault(msg, 16, ""),
    attachPathList: jspb.Message.getRepeatedField(msg, 17),
    attachInfoMap: (f = msg.getAttachInfoMap()) ? f.toObject(includeInstance, undefined) : [],
    replicaSetsList: jspb.Message.toObjectList(msg.getReplicaSetsList(),
    proto.openstorage.api.ReplicaSet.toObject, includeInstance),
    runtimeStateList: jspb.Message.toObjectList(msg.getRuntimeStateList(),
    proto.openstorage.api.RuntimeStateMap.toObject, includeInstance),
    error: jspb.Message.getFieldWithDefault(msg, 21, ""),
    volumeConsumersList: jspb.Message.toObjectList(msg.getVolumeConsumersList(),
    proto.openstorage.api.VolumeConsumer.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.Volume}
 */
proto.openstorage.api.Volume.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.Volume;
  return proto.openstorage.api.Volume.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.Volume} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.Volume}
 */
proto.openstorage.api.Volume.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = new proto.openstorage.api.Source;
      reader.readMessage(value,proto.openstorage.api.Source.deserializeBinaryFromReader);
      msg.setSource(value);
      break;
    case 3:
      var value = new proto.openstorage.api.Group;
      reader.readMessage(value,proto.openstorage.api.Group.deserializeBinaryFromReader);
      msg.setGroup(value);
      break;
    case 4:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setReadonly(value);
      break;
    case 5:
      var value = new proto.openstorage.api.VolumeLocator;
      reader.readMessage(value,proto.openstorage.api.VolumeLocator.deserializeBinaryFromReader);
      msg.setLocator(value);
      break;
    case 6:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setCtime(value);
      break;
    case 7:
      var value = new proto.openstorage.api.VolumeSpec;
      reader.readMessage(value,proto.openstorage.api.VolumeSpec.deserializeBinaryFromReader);
      msg.setSpec(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setUsage(value);
      break;
    case 9:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setLastScan(value);
      break;
    case 10:
      var value = /** @type {!proto.openstorage.api.FSType} */ (reader.readEnum());
      msg.setFormat(value);
      break;
    case 11:
      var value = /** @type {!proto.openstorage.api.VolumeStatus} */ (reader.readEnum());
      msg.setStatus(value);
      break;
    case 12:
      var value = /** @type {!proto.openstorage.api.VolumeState} */ (reader.readEnum());
      msg.setState(value);
      break;
    case 13:
      var value = /** @type {string} */ (reader.readString());
      msg.setAttachedOn(value);
      break;
    case 14:
      var value = /** @type {!proto.openstorage.api.AttachState} */ (reader.readEnum());
      msg.setAttachedState(value);
      break;
    case 15:
      var value = /** @type {string} */ (reader.readString());
      msg.setDevicePath(value);
      break;
    case 16:
      var value = /** @type {string} */ (reader.readString());
      msg.setSecureDevicePath(value);
      break;
    case 17:
      var value = /** @type {string} */ (reader.readString());
      msg.addAttachPath(value);
      break;
    case 18:
      var value = msg.getAttachInfoMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString);
         });
      break;
    case 19:
      var value = new proto.openstorage.api.ReplicaSet;
      reader.readMessage(value,proto.openstorage.api.ReplicaSet.deserializeBinaryFromReader);
      msg.addReplicaSets(value);
      break;
    case 20:
      var value = new proto.openstorage.api.RuntimeStateMap;
      reader.readMessage(value,proto.openstorage.api.RuntimeStateMap.deserializeBinaryFromReader);
      msg.addRuntimeState(value);
      break;
    case 21:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
      break;
    case 22:
      var value = new proto.openstorage.api.VolumeConsumer;
      reader.readMessage(value,proto.openstorage.api.VolumeConsumer.deserializeBinaryFromReader);
      msg.addVolumeConsumers(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.Volume.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.Volume.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.Volume} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Volume.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSource();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.openstorage.api.Source.serializeBinaryToWriter
    );
  }
  f = message.getGroup();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.openstorage.api.Group.serializeBinaryToWriter
    );
  }
  f = message.getReadonly();
  if (f) {
    writer.writeBool(
      4,
      f
    );
  }
  f = message.getLocator();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      proto.openstorage.api.VolumeLocator.serializeBinaryToWriter
    );
  }
  f = message.getCtime();
  if (f != null) {
    writer.writeMessage(
      6,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getSpec();
  if (f != null) {
    writer.writeMessage(
      7,
      f,
      proto.openstorage.api.VolumeSpec.serializeBinaryToWriter
    );
  }
  f = message.getUsage();
  if (f !== 0) {
    writer.writeUint64(
      8,
      f
    );
  }
  f = message.getLastScan();
  if (f != null) {
    writer.writeMessage(
      9,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getFormat();
  if (f !== 0.0) {
    writer.writeEnum(
      10,
      f
    );
  }
  f = message.getStatus();
  if (f !== 0.0) {
    writer.writeEnum(
      11,
      f
    );
  }
  f = message.getState();
  if (f !== 0.0) {
    writer.writeEnum(
      12,
      f
    );
  }
  f = message.getAttachedOn();
  if (f.length > 0) {
    writer.writeString(
      13,
      f
    );
  }
  f = message.getAttachedState();
  if (f !== 0.0) {
    writer.writeEnum(
      14,
      f
    );
  }
  f = message.getDevicePath();
  if (f.length > 0) {
    writer.writeString(
      15,
      f
    );
  }
  f = message.getSecureDevicePath();
  if (f.length > 0) {
    writer.writeString(
      16,
      f
    );
  }
  f = message.getAttachPathList();
  if (f.length > 0) {
    writer.writeRepeatedString(
      17,
      f
    );
  }
  f = message.getAttachInfoMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(18, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
  f = message.getReplicaSetsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      19,
      f,
      proto.openstorage.api.ReplicaSet.serializeBinaryToWriter
    );
  }
  f = message.getRuntimeStateList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      20,
      f,
      proto.openstorage.api.RuntimeStateMap.serializeBinaryToWriter
    );
  }
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      21,
      f
    );
  }
  f = message.getVolumeConsumersList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      22,
      f,
      proto.openstorage.api.VolumeConsumer.serializeBinaryToWriter
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.openstorage.api.Volume.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.Volume.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional Source source = 2;
 * @return {?proto.openstorage.api.Source}
 */
proto.openstorage.api.Volume.prototype.getSource = function() {
  return /** @type{?proto.openstorage.api.Source} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.Source, 2));
};


/** @param {?proto.openstorage.api.Source|undefined} value */
proto.openstorage.api.Volume.prototype.setSource = function(value) {
  jspb.Message.setWrapperField(this, 2, value);
};


proto.openstorage.api.Volume.prototype.clearSource = function() {
  this.setSource(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.Volume.prototype.hasSource = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional Group group = 3;
 * @return {?proto.openstorage.api.Group}
 */
proto.openstorage.api.Volume.prototype.getGroup = function() {
  return /** @type{?proto.openstorage.api.Group} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.Group, 3));
};


/** @param {?proto.openstorage.api.Group|undefined} value */
proto.openstorage.api.Volume.prototype.setGroup = function(value) {
  jspb.Message.setWrapperField(this, 3, value);
};


proto.openstorage.api.Volume.prototype.clearGroup = function() {
  this.setGroup(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.Volume.prototype.hasGroup = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * optional bool readonly = 4;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.Volume.prototype.getReadonly = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 4, false));
};


/** @param {boolean} value */
proto.openstorage.api.Volume.prototype.setReadonly = function(value) {
  jspb.Message.setField(this, 4, value);
};


/**
 * optional VolumeLocator locator = 5;
 * @return {?proto.openstorage.api.VolumeLocator}
 */
proto.openstorage.api.Volume.prototype.getLocator = function() {
  return /** @type{?proto.openstorage.api.VolumeLocator} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeLocator, 5));
};


/** @param {?proto.openstorage.api.VolumeLocator|undefined} value */
proto.openstorage.api.Volume.prototype.setLocator = function(value) {
  jspb.Message.setWrapperField(this, 5, value);
};


proto.openstorage.api.Volume.prototype.clearLocator = function() {
  this.setLocator(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.Volume.prototype.hasLocator = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional google.protobuf.Timestamp ctime = 6;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.openstorage.api.Volume.prototype.getCtime = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 6));
};


/** @param {?proto.google.protobuf.Timestamp|undefined} value */
proto.openstorage.api.Volume.prototype.setCtime = function(value) {
  jspb.Message.setWrapperField(this, 6, value);
};


proto.openstorage.api.Volume.prototype.clearCtime = function() {
  this.setCtime(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.Volume.prototype.hasCtime = function() {
  return jspb.Message.getField(this, 6) != null;
};


/**
 * optional VolumeSpec spec = 7;
 * @return {?proto.openstorage.api.VolumeSpec}
 */
proto.openstorage.api.Volume.prototype.getSpec = function() {
  return /** @type{?proto.openstorage.api.VolumeSpec} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeSpec, 7));
};


/** @param {?proto.openstorage.api.VolumeSpec|undefined} value */
proto.openstorage.api.Volume.prototype.setSpec = function(value) {
  jspb.Message.setWrapperField(this, 7, value);
};


proto.openstorage.api.Volume.prototype.clearSpec = function() {
  this.setSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.Volume.prototype.hasSpec = function() {
  return jspb.Message.getField(this, 7) != null;
};


/**
 * optional uint64 usage = 8;
 * @return {number}
 */
proto.openstorage.api.Volume.prototype.getUsage = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/** @param {number} value */
proto.openstorage.api.Volume.prototype.setUsage = function(value) {
  jspb.Message.setField(this, 8, value);
};


/**
 * optional google.protobuf.Timestamp last_scan = 9;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.openstorage.api.Volume.prototype.getLastScan = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 9));
};


/** @param {?proto.google.protobuf.Timestamp|undefined} value */
proto.openstorage.api.Volume.prototype.setLastScan = function(value) {
  jspb.Message.setWrapperField(this, 9, value);
};


proto.openstorage.api.Volume.prototype.clearLastScan = function() {
  this.setLastScan(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.Volume.prototype.hasLastScan = function() {
  return jspb.Message.getField(this, 9) != null;
};


/**
 * optional FSType format = 10;
 * @return {!proto.openstorage.api.FSType}
 */
proto.openstorage.api.Volume.prototype.getFormat = function() {
  return /** @type {!proto.openstorage.api.FSType} */ (jspb.Message.getFieldWithDefault(this, 10, 0));
};


/** @param {!proto.openstorage.api.FSType} value */
proto.openstorage.api.Volume.prototype.setFormat = function(value) {
  jspb.Message.setField(this, 10, value);
};


/**
 * optional VolumeStatus status = 11;
 * @return {!proto.openstorage.api.VolumeStatus}
 */
proto.openstorage.api.Volume.prototype.getStatus = function() {
  return /** @type {!proto.openstorage.api.VolumeStatus} */ (jspb.Message.getFieldWithDefault(this, 11, 0));
};


/** @param {!proto.openstorage.api.VolumeStatus} value */
proto.openstorage.api.Volume.prototype.setStatus = function(value) {
  jspb.Message.setField(this, 11, value);
};


/**
 * optional VolumeState state = 12;
 * @return {!proto.openstorage.api.VolumeState}
 */
proto.openstorage.api.Volume.prototype.getState = function() {
  return /** @type {!proto.openstorage.api.VolumeState} */ (jspb.Message.getFieldWithDefault(this, 12, 0));
};


/** @param {!proto.openstorage.api.VolumeState} value */
proto.openstorage.api.Volume.prototype.setState = function(value) {
  jspb.Message.setField(this, 12, value);
};


/**
 * optional string attached_on = 13;
 * @return {string}
 */
proto.openstorage.api.Volume.prototype.getAttachedOn = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 13, ""));
};


/** @param {string} value */
proto.openstorage.api.Volume.prototype.setAttachedOn = function(value) {
  jspb.Message.setField(this, 13, value);
};


/**
 * optional AttachState attached_state = 14;
 * @return {!proto.openstorage.api.AttachState}
 */
proto.openstorage.api.Volume.prototype.getAttachedState = function() {
  return /** @type {!proto.openstorage.api.AttachState} */ (jspb.Message.getFieldWithDefault(this, 14, 0));
};


/** @param {!proto.openstorage.api.AttachState} value */
proto.openstorage.api.Volume.prototype.setAttachedState = function(value) {
  jspb.Message.setField(this, 14, value);
};


/**
 * optional string device_path = 15;
 * @return {string}
 */
proto.openstorage.api.Volume.prototype.getDevicePath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 15, ""));
};


/** @param {string} value */
proto.openstorage.api.Volume.prototype.setDevicePath = function(value) {
  jspb.Message.setField(this, 15, value);
};


/**
 * optional string secure_device_path = 16;
 * @return {string}
 */
proto.openstorage.api.Volume.prototype.getSecureDevicePath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 16, ""));
};


/** @param {string} value */
proto.openstorage.api.Volume.prototype.setSecureDevicePath = function(value) {
  jspb.Message.setField(this, 16, value);
};


/**
 * repeated string attach_path = 17;
 * @return {!Array.<string>}
 */
proto.openstorage.api.Volume.prototype.getAttachPathList = function() {
  return /** @type {!Array.<string>} */ (jspb.Message.getRepeatedField(this, 17));
};


/** @param {!Array.<string>} value */
proto.openstorage.api.Volume.prototype.setAttachPathList = function(value) {
  jspb.Message.setField(this, 17, value || []);
};


/**
 * @param {!string} value
 * @param {number=} opt_index
 */
proto.openstorage.api.Volume.prototype.addAttachPath = function(value, opt_index) {
  jspb.Message.addToRepeatedField(this, 17, value, opt_index);
};


proto.openstorage.api.Volume.prototype.clearAttachPathList = function() {
  this.setAttachPathList([]);
};


/**
 * map<string, string> attach_info = 18;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.openstorage.api.Volume.prototype.getAttachInfoMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 18, opt_noLazyCreate,
      null));
};


proto.openstorage.api.Volume.prototype.clearAttachInfoMap = function() {
  this.getAttachInfoMap().clear();
};


/**
 * repeated ReplicaSet replica_sets = 19;
 * @return {!Array.<!proto.openstorage.api.ReplicaSet>}
 */
proto.openstorage.api.Volume.prototype.getReplicaSetsList = function() {
  return /** @type{!Array.<!proto.openstorage.api.ReplicaSet>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.openstorage.api.ReplicaSet, 19));
};


/** @param {!Array.<!proto.openstorage.api.ReplicaSet>} value */
proto.openstorage.api.Volume.prototype.setReplicaSetsList = function(value) {
  jspb.Message.setRepeatedWrapperField(this, 19, value);
};


/**
 * @param {!proto.openstorage.api.ReplicaSet=} opt_value
 * @param {number=} opt_index
 * @return {!proto.openstorage.api.ReplicaSet}
 */
proto.openstorage.api.Volume.prototype.addReplicaSets = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 19, opt_value, proto.openstorage.api.ReplicaSet, opt_index);
};


proto.openstorage.api.Volume.prototype.clearReplicaSetsList = function() {
  this.setReplicaSetsList([]);
};


/**
 * repeated RuntimeStateMap runtime_state = 20;
 * @return {!Array.<!proto.openstorage.api.RuntimeStateMap>}
 */
proto.openstorage.api.Volume.prototype.getRuntimeStateList = function() {
  return /** @type{!Array.<!proto.openstorage.api.RuntimeStateMap>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.openstorage.api.RuntimeStateMap, 20));
};


/** @param {!Array.<!proto.openstorage.api.RuntimeStateMap>} value */
proto.openstorage.api.Volume.prototype.setRuntimeStateList = function(value) {
  jspb.Message.setRepeatedWrapperField(this, 20, value);
};


/**
 * @param {!proto.openstorage.api.RuntimeStateMap=} opt_value
 * @param {number=} opt_index
 * @return {!proto.openstorage.api.RuntimeStateMap}
 */
proto.openstorage.api.Volume.prototype.addRuntimeState = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 20, opt_value, proto.openstorage.api.RuntimeStateMap, opt_index);
};


proto.openstorage.api.Volume.prototype.clearRuntimeStateList = function() {
  this.setRuntimeStateList([]);
};


/**
 * optional string error = 21;
 * @return {string}
 */
proto.openstorage.api.Volume.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 21, ""));
};


/** @param {string} value */
proto.openstorage.api.Volume.prototype.setError = function(value) {
  jspb.Message.setField(this, 21, value);
};


/**
 * repeated VolumeConsumer volume_consumers = 22;
 * @return {!Array.<!proto.openstorage.api.VolumeConsumer>}
 */
proto.openstorage.api.Volume.prototype.getVolumeConsumersList = function() {
  return /** @type{!Array.<!proto.openstorage.api.VolumeConsumer>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.openstorage.api.VolumeConsumer, 22));
};


/** @param {!Array.<!proto.openstorage.api.VolumeConsumer>} value */
proto.openstorage.api.Volume.prototype.setVolumeConsumersList = function(value) {
  jspb.Message.setRepeatedWrapperField(this, 22, value);
};


/**
 * @param {!proto.openstorage.api.VolumeConsumer=} opt_value
 * @param {number=} opt_index
 * @return {!proto.openstorage.api.VolumeConsumer}
 */
proto.openstorage.api.Volume.prototype.addVolumeConsumers = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 22, opt_value, proto.openstorage.api.VolumeConsumer, opt_index);
};


proto.openstorage.api.Volume.prototype.clearVolumeConsumersList = function() {
  this.setVolumeConsumersList([]);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.Stats = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.Stats, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.Stats.displayName = 'proto.openstorage.api.Stats';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.Stats.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.Stats.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.Stats} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Stats.toObject = function(includeInstance, msg) {
  var f, obj = {
    reads: jspb.Message.getFieldWithDefault(msg, 1, 0),
    readMs: jspb.Message.getFieldWithDefault(msg, 2, 0),
    readBytes: jspb.Message.getFieldWithDefault(msg, 3, 0),
    writes: jspb.Message.getFieldWithDefault(msg, 4, 0),
    writeMs: jspb.Message.getFieldWithDefault(msg, 5, 0),
    writeBytes: jspb.Message.getFieldWithDefault(msg, 6, 0),
    ioProgress: jspb.Message.getFieldWithDefault(msg, 7, 0),
    ioMs: jspb.Message.getFieldWithDefault(msg, 8, 0),
    bytesUsed: jspb.Message.getFieldWithDefault(msg, 9, 0),
    intervalMs: jspb.Message.getFieldWithDefault(msg, 10, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.Stats}
 */
proto.openstorage.api.Stats.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.Stats;
  return proto.openstorage.api.Stats.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.Stats} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.Stats}
 */
proto.openstorage.api.Stats.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setReads(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setReadMs(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setReadBytes(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setWrites(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setWriteMs(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setWriteBytes(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setIoProgress(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setIoMs(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setBytesUsed(value);
      break;
    case 10:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setIntervalMs(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.Stats.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.Stats.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.Stats} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Stats.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getReads();
  if (f !== 0) {
    writer.writeUint64(
      1,
      f
    );
  }
  f = message.getReadMs();
  if (f !== 0) {
    writer.writeUint64(
      2,
      f
    );
  }
  f = message.getReadBytes();
  if (f !== 0) {
    writer.writeUint64(
      3,
      f
    );
  }
  f = message.getWrites();
  if (f !== 0) {
    writer.writeUint64(
      4,
      f
    );
  }
  f = message.getWriteMs();
  if (f !== 0) {
    writer.writeUint64(
      5,
      f
    );
  }
  f = message.getWriteBytes();
  if (f !== 0) {
    writer.writeUint64(
      6,
      f
    );
  }
  f = message.getIoProgress();
  if (f !== 0) {
    writer.writeUint64(
      7,
      f
    );
  }
  f = message.getIoMs();
  if (f !== 0) {
    writer.writeUint64(
      8,
      f
    );
  }
  f = message.getBytesUsed();
  if (f !== 0) {
    writer.writeUint64(
      9,
      f
    );
  }
  f = message.getIntervalMs();
  if (f !== 0) {
    writer.writeUint64(
      10,
      f
    );
  }
};


/**
 * optional uint64 reads = 1;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getReads = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setReads = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional uint64 read_ms = 2;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getReadMs = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setReadMs = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional uint64 read_bytes = 3;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getReadBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setReadBytes = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * optional uint64 writes = 4;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getWrites = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setWrites = function(value) {
  jspb.Message.setField(this, 4, value);
};


/**
 * optional uint64 write_ms = 5;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getWriteMs = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setWriteMs = function(value) {
  jspb.Message.setField(this, 5, value);
};


/**
 * optional uint64 write_bytes = 6;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getWriteBytes = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setWriteBytes = function(value) {
  jspb.Message.setField(this, 6, value);
};


/**
 * optional uint64 io_progress = 7;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getIoProgress = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setIoProgress = function(value) {
  jspb.Message.setField(this, 7, value);
};


/**
 * optional uint64 io_ms = 8;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getIoMs = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setIoMs = function(value) {
  jspb.Message.setField(this, 8, value);
};


/**
 * optional uint64 bytes_used = 9;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getBytesUsed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setBytesUsed = function(value) {
  jspb.Message.setField(this, 9, value);
};


/**
 * optional uint64 interval_ms = 10;
 * @return {number}
 */
proto.openstorage.api.Stats.prototype.getIntervalMs = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 10, 0));
};


/** @param {number} value */
proto.openstorage.api.Stats.prototype.setIntervalMs = function(value) {
  jspb.Message.setField(this, 10, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.Alert = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.Alert, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.Alert.displayName = 'proto.openstorage.api.Alert';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.Alert.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.Alert.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.Alert} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Alert.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, 0),
    severity: jspb.Message.getFieldWithDefault(msg, 2, 0),
    alertType: jspb.Message.getFieldWithDefault(msg, 3, 0),
    message: jspb.Message.getFieldWithDefault(msg, 4, ""),
    timestamp: (f = msg.getTimestamp()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    resourceId: jspb.Message.getFieldWithDefault(msg, 6, ""),
    resource: jspb.Message.getFieldWithDefault(msg, 7, 0),
    cleared: jspb.Message.getFieldWithDefault(msg, 8, false),
    ttl: jspb.Message.getFieldWithDefault(msg, 9, 0),
    uniqueTag: jspb.Message.getFieldWithDefault(msg, 10, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.Alert}
 */
proto.openstorage.api.Alert.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.Alert;
  return proto.openstorage.api.Alert.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.Alert} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.Alert}
 */
proto.openstorage.api.Alert.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {!proto.openstorage.api.SeverityType} */ (reader.readEnum());
      msg.setSeverity(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setAlertType(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setMessage(value);
      break;
    case 5:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setTimestamp(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setResourceId(value);
      break;
    case 7:
      var value = /** @type {!proto.openstorage.api.ResourceType} */ (reader.readEnum());
      msg.setResource(value);
      break;
    case 8:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setCleared(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setTtl(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.setUniqueTag(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.Alert.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.Alert.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.Alert} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Alert.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getSeverity();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getAlertType();
  if (f !== 0) {
    writer.writeInt64(
      3,
      f
    );
  }
  f = message.getMessage();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getTimestamp();
  if (f != null) {
    writer.writeMessage(
      5,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getResourceId();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getResource();
  if (f !== 0.0) {
    writer.writeEnum(
      7,
      f
    );
  }
  f = message.getCleared();
  if (f) {
    writer.writeBool(
      8,
      f
    );
  }
  f = message.getTtl();
  if (f !== 0) {
    writer.writeUint64(
      9,
      f
    );
  }
  f = message.getUniqueTag();
  if (f.length > 0) {
    writer.writeString(
      10,
      f
    );
  }
};


/**
 * optional int64 id = 1;
 * @return {number}
 */
proto.openstorage.api.Alert.prototype.getId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {number} value */
proto.openstorage.api.Alert.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional SeverityType severity = 2;
 * @return {!proto.openstorage.api.SeverityType}
 */
proto.openstorage.api.Alert.prototype.getSeverity = function() {
  return /** @type {!proto.openstorage.api.SeverityType} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/** @param {!proto.openstorage.api.SeverityType} value */
proto.openstorage.api.Alert.prototype.setSeverity = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional int64 alert_type = 3;
 * @return {number}
 */
proto.openstorage.api.Alert.prototype.getAlertType = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/** @param {number} value */
proto.openstorage.api.Alert.prototype.setAlertType = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * optional string message = 4;
 * @return {string}
 */
proto.openstorage.api.Alert.prototype.getMessage = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/** @param {string} value */
proto.openstorage.api.Alert.prototype.setMessage = function(value) {
  jspb.Message.setField(this, 4, value);
};


/**
 * optional google.protobuf.Timestamp timestamp = 5;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.openstorage.api.Alert.prototype.getTimestamp = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 5));
};


/** @param {?proto.google.protobuf.Timestamp|undefined} value */
proto.openstorage.api.Alert.prototype.setTimestamp = function(value) {
  jspb.Message.setWrapperField(this, 5, value);
};


proto.openstorage.api.Alert.prototype.clearTimestamp = function() {
  this.setTimestamp(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.Alert.prototype.hasTimestamp = function() {
  return jspb.Message.getField(this, 5) != null;
};


/**
 * optional string resource_id = 6;
 * @return {string}
 */
proto.openstorage.api.Alert.prototype.getResourceId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/** @param {string} value */
proto.openstorage.api.Alert.prototype.setResourceId = function(value) {
  jspb.Message.setField(this, 6, value);
};


/**
 * optional ResourceType resource = 7;
 * @return {!proto.openstorage.api.ResourceType}
 */
proto.openstorage.api.Alert.prototype.getResource = function() {
  return /** @type {!proto.openstorage.api.ResourceType} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/** @param {!proto.openstorage.api.ResourceType} value */
proto.openstorage.api.Alert.prototype.setResource = function(value) {
  jspb.Message.setField(this, 7, value);
};


/**
 * optional bool cleared = 8;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.Alert.prototype.getCleared = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 8, false));
};


/** @param {boolean} value */
proto.openstorage.api.Alert.prototype.setCleared = function(value) {
  jspb.Message.setField(this, 8, value);
};


/**
 * optional uint64 ttl = 9;
 * @return {number}
 */
proto.openstorage.api.Alert.prototype.getTtl = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/** @param {number} value */
proto.openstorage.api.Alert.prototype.setTtl = function(value) {
  jspb.Message.setField(this, 9, value);
};


/**
 * optional string unique_tag = 10;
 * @return {string}
 */
proto.openstorage.api.Alert.prototype.getUniqueTag = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 10, ""));
};


/** @param {string} value */
proto.openstorage.api.Alert.prototype.setUniqueTag = function(value) {
  jspb.Message.setField(this, 10, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.Alerts = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.openstorage.api.Alerts.repeatedFields_, null);
};
goog.inherits(proto.openstorage.api.Alerts, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.Alerts.displayName = 'proto.openstorage.api.Alerts';
}
/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.openstorage.api.Alerts.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.Alerts.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.Alerts.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.Alerts} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Alerts.toObject = function(includeInstance, msg) {
  var f, obj = {
    alertList: jspb.Message.toObjectList(msg.getAlertList(),
    proto.openstorage.api.Alert.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.Alerts}
 */
proto.openstorage.api.Alerts.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.Alerts;
  return proto.openstorage.api.Alerts.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.Alerts} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.Alerts}
 */
proto.openstorage.api.Alerts.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.Alert;
      reader.readMessage(value,proto.openstorage.api.Alert.deserializeBinaryFromReader);
      msg.addAlert(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.Alerts.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.Alerts.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.Alerts} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.Alerts.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAlertList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.openstorage.api.Alert.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Alert alert = 1;
 * @return {!Array.<!proto.openstorage.api.Alert>}
 */
proto.openstorage.api.Alerts.prototype.getAlertList = function() {
  return /** @type{!Array.<!proto.openstorage.api.Alert>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.openstorage.api.Alert, 1));
};


/** @param {!Array.<!proto.openstorage.api.Alert>} value */
proto.openstorage.api.Alerts.prototype.setAlertList = function(value) {
  jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.openstorage.api.Alert=} opt_value
 * @param {number=} opt_index
 * @return {!proto.openstorage.api.Alert}
 */
proto.openstorage.api.Alerts.prototype.addAlert = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.openstorage.api.Alert, opt_index);
};


proto.openstorage.api.Alerts.prototype.clearAlertList = function() {
  this.setAlertList([]);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeCreateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeCreateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeCreateRequest.displayName = 'proto.openstorage.api.VolumeCreateRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeCreateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeCreateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeCreateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeCreateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    locator: (f = msg.getLocator()) && proto.openstorage.api.VolumeLocator.toObject(includeInstance, f),
    source: (f = msg.getSource()) && proto.openstorage.api.Source.toObject(includeInstance, f),
    spec: (f = msg.getSpec()) && proto.openstorage.api.VolumeSpec.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeCreateRequest}
 */
proto.openstorage.api.VolumeCreateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeCreateRequest;
  return proto.openstorage.api.VolumeCreateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeCreateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeCreateRequest}
 */
proto.openstorage.api.VolumeCreateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.VolumeLocator;
      reader.readMessage(value,proto.openstorage.api.VolumeLocator.deserializeBinaryFromReader);
      msg.setLocator(value);
      break;
    case 2:
      var value = new proto.openstorage.api.Source;
      reader.readMessage(value,proto.openstorage.api.Source.deserializeBinaryFromReader);
      msg.setSource(value);
      break;
    case 3:
      var value = new proto.openstorage.api.VolumeSpec;
      reader.readMessage(value,proto.openstorage.api.VolumeSpec.deserializeBinaryFromReader);
      msg.setSpec(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeCreateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeCreateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeCreateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeCreateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getLocator();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.VolumeLocator.serializeBinaryToWriter
    );
  }
  f = message.getSource();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.openstorage.api.Source.serializeBinaryToWriter
    );
  }
  f = message.getSpec();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.openstorage.api.VolumeSpec.serializeBinaryToWriter
    );
  }
};


/**
 * optional VolumeLocator locator = 1;
 * @return {?proto.openstorage.api.VolumeLocator}
 */
proto.openstorage.api.VolumeCreateRequest.prototype.getLocator = function() {
  return /** @type{?proto.openstorage.api.VolumeLocator} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeLocator, 1));
};


/** @param {?proto.openstorage.api.VolumeLocator|undefined} value */
proto.openstorage.api.VolumeCreateRequest.prototype.setLocator = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.VolumeCreateRequest.prototype.clearLocator = function() {
  this.setLocator(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeCreateRequest.prototype.hasLocator = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional Source source = 2;
 * @return {?proto.openstorage.api.Source}
 */
proto.openstorage.api.VolumeCreateRequest.prototype.getSource = function() {
  return /** @type{?proto.openstorage.api.Source} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.Source, 2));
};


/** @param {?proto.openstorage.api.Source|undefined} value */
proto.openstorage.api.VolumeCreateRequest.prototype.setSource = function(value) {
  jspb.Message.setWrapperField(this, 2, value);
};


proto.openstorage.api.VolumeCreateRequest.prototype.clearSource = function() {
  this.setSource(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeCreateRequest.prototype.hasSource = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional VolumeSpec spec = 3;
 * @return {?proto.openstorage.api.VolumeSpec}
 */
proto.openstorage.api.VolumeCreateRequest.prototype.getSpec = function() {
  return /** @type{?proto.openstorage.api.VolumeSpec} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeSpec, 3));
};


/** @param {?proto.openstorage.api.VolumeSpec|undefined} value */
proto.openstorage.api.VolumeCreateRequest.prototype.setSpec = function(value) {
  jspb.Message.setWrapperField(this, 3, value);
};


proto.openstorage.api.VolumeCreateRequest.prototype.clearSpec = function() {
  this.setSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeCreateRequest.prototype.hasSpec = function() {
  return jspb.Message.getField(this, 3) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeResponse.displayName = 'proto.openstorage.api.VolumeResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    error: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeResponse}
 */
proto.openstorage.api.VolumeResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeResponse;
  return proto.openstorage.api.VolumeResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeResponse}
 */
proto.openstorage.api.VolumeResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string error = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeResponse.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeResponse.prototype.setError = function(value) {
  jspb.Message.setField(this, 1, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeCreateResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeCreateResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeCreateResponse.displayName = 'proto.openstorage.api.VolumeCreateResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeCreateResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeCreateResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeCreateResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeCreateResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    volumeResponse: (f = msg.getVolumeResponse()) && proto.openstorage.api.VolumeResponse.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeCreateResponse}
 */
proto.openstorage.api.VolumeCreateResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeCreateResponse;
  return proto.openstorage.api.VolumeCreateResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeCreateResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeCreateResponse}
 */
proto.openstorage.api.VolumeCreateResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = new proto.openstorage.api.VolumeResponse;
      reader.readMessage(value,proto.openstorage.api.VolumeResponse.deserializeBinaryFromReader);
      msg.setVolumeResponse(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeCreateResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeCreateResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeCreateResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeCreateResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getVolumeResponse();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.openstorage.api.VolumeResponse.serializeBinaryToWriter
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeCreateResponse.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeCreateResponse.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional VolumeResponse volume_response = 2;
 * @return {?proto.openstorage.api.VolumeResponse}
 */
proto.openstorage.api.VolumeCreateResponse.prototype.getVolumeResponse = function() {
  return /** @type{?proto.openstorage.api.VolumeResponse} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeResponse, 2));
};


/** @param {?proto.openstorage.api.VolumeResponse|undefined} value */
proto.openstorage.api.VolumeCreateResponse.prototype.setVolumeResponse = function(value) {
  jspb.Message.setWrapperField(this, 2, value);
};


proto.openstorage.api.VolumeCreateResponse.prototype.clearVolumeResponse = function() {
  this.setVolumeResponse(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeCreateResponse.prototype.hasVolumeResponse = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeStateAction = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeStateAction, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeStateAction.displayName = 'proto.openstorage.api.VolumeStateAction';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeStateAction.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeStateAction.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeStateAction} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeStateAction.toObject = function(includeInstance, msg) {
  var f, obj = {
    attach: jspb.Message.getFieldWithDefault(msg, 1, 0),
    mount: jspb.Message.getFieldWithDefault(msg, 2, 0),
    mountPath: jspb.Message.getFieldWithDefault(msg, 3, ""),
    devicePath: jspb.Message.getFieldWithDefault(msg, 4, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeStateAction}
 */
proto.openstorage.api.VolumeStateAction.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeStateAction;
  return proto.openstorage.api.VolumeStateAction.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeStateAction} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeStateAction}
 */
proto.openstorage.api.VolumeStateAction.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!proto.openstorage.api.VolumeActionParam} */ (reader.readEnum());
      msg.setAttach(value);
      break;
    case 2:
      var value = /** @type {!proto.openstorage.api.VolumeActionParam} */ (reader.readEnum());
      msg.setMount(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setMountPath(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setDevicePath(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeStateAction.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeStateAction.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeStateAction} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeStateAction.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAttach();
  if (f !== 0.0) {
    writer.writeEnum(
      1,
      f
    );
  }
  f = message.getMount();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
  f = message.getMountPath();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getDevicePath();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
};


/**
 * optional VolumeActionParam attach = 1;
 * @return {!proto.openstorage.api.VolumeActionParam}
 */
proto.openstorage.api.VolumeStateAction.prototype.getAttach = function() {
  return /** @type {!proto.openstorage.api.VolumeActionParam} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {!proto.openstorage.api.VolumeActionParam} value */
proto.openstorage.api.VolumeStateAction.prototype.setAttach = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional VolumeActionParam mount = 2;
 * @return {!proto.openstorage.api.VolumeActionParam}
 */
proto.openstorage.api.VolumeStateAction.prototype.getMount = function() {
  return /** @type {!proto.openstorage.api.VolumeActionParam} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/** @param {!proto.openstorage.api.VolumeActionParam} value */
proto.openstorage.api.VolumeStateAction.prototype.setMount = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional string mount_path = 3;
 * @return {string}
 */
proto.openstorage.api.VolumeStateAction.prototype.getMountPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeStateAction.prototype.setMountPath = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * optional string device_path = 4;
 * @return {string}
 */
proto.openstorage.api.VolumeStateAction.prototype.getDevicePath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeStateAction.prototype.setDevicePath = function(value) {
  jspb.Message.setField(this, 4, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeSetRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeSetRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeSetRequest.displayName = 'proto.openstorage.api.VolumeSetRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeSetRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeSetRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeSetRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeSetRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    locator: (f = msg.getLocator()) && proto.openstorage.api.VolumeLocator.toObject(includeInstance, f),
    spec: (f = msg.getSpec()) && proto.openstorage.api.VolumeSpec.toObject(includeInstance, f),
    action: (f = msg.getAction()) && proto.openstorage.api.VolumeStateAction.toObject(includeInstance, f),
    optionsMap: (f = msg.getOptionsMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeSetRequest}
 */
proto.openstorage.api.VolumeSetRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeSetRequest;
  return proto.openstorage.api.VolumeSetRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeSetRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeSetRequest}
 */
proto.openstorage.api.VolumeSetRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.VolumeLocator;
      reader.readMessage(value,proto.openstorage.api.VolumeLocator.deserializeBinaryFromReader);
      msg.setLocator(value);
      break;
    case 2:
      var value = new proto.openstorage.api.VolumeSpec;
      reader.readMessage(value,proto.openstorage.api.VolumeSpec.deserializeBinaryFromReader);
      msg.setSpec(value);
      break;
    case 3:
      var value = new proto.openstorage.api.VolumeStateAction;
      reader.readMessage(value,proto.openstorage.api.VolumeStateAction.deserializeBinaryFromReader);
      msg.setAction(value);
      break;
    case 4:
      var value = msg.getOptionsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeSetRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeSetRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeSetRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeSetRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getLocator();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.VolumeLocator.serializeBinaryToWriter
    );
  }
  f = message.getSpec();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.openstorage.api.VolumeSpec.serializeBinaryToWriter
    );
  }
  f = message.getAction();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.openstorage.api.VolumeStateAction.serializeBinaryToWriter
    );
  }
  f = message.getOptionsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(4, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * optional VolumeLocator locator = 1;
 * @return {?proto.openstorage.api.VolumeLocator}
 */
proto.openstorage.api.VolumeSetRequest.prototype.getLocator = function() {
  return /** @type{?proto.openstorage.api.VolumeLocator} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeLocator, 1));
};


/** @param {?proto.openstorage.api.VolumeLocator|undefined} value */
proto.openstorage.api.VolumeSetRequest.prototype.setLocator = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.VolumeSetRequest.prototype.clearLocator = function() {
  this.setLocator(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeSetRequest.prototype.hasLocator = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional VolumeSpec spec = 2;
 * @return {?proto.openstorage.api.VolumeSpec}
 */
proto.openstorage.api.VolumeSetRequest.prototype.getSpec = function() {
  return /** @type{?proto.openstorage.api.VolumeSpec} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeSpec, 2));
};


/** @param {?proto.openstorage.api.VolumeSpec|undefined} value */
proto.openstorage.api.VolumeSetRequest.prototype.setSpec = function(value) {
  jspb.Message.setWrapperField(this, 2, value);
};


proto.openstorage.api.VolumeSetRequest.prototype.clearSpec = function() {
  this.setSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeSetRequest.prototype.hasSpec = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional VolumeStateAction action = 3;
 * @return {?proto.openstorage.api.VolumeStateAction}
 */
proto.openstorage.api.VolumeSetRequest.prototype.getAction = function() {
  return /** @type{?proto.openstorage.api.VolumeStateAction} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeStateAction, 3));
};


/** @param {?proto.openstorage.api.VolumeStateAction|undefined} value */
proto.openstorage.api.VolumeSetRequest.prototype.setAction = function(value) {
  jspb.Message.setWrapperField(this, 3, value);
};


proto.openstorage.api.VolumeSetRequest.prototype.clearAction = function() {
  this.setAction(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeSetRequest.prototype.hasAction = function() {
  return jspb.Message.getField(this, 3) != null;
};


/**
 * map<string, string> options = 4;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.openstorage.api.VolumeSetRequest.prototype.getOptionsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 4, opt_noLazyCreate,
      null));
};


proto.openstorage.api.VolumeSetRequest.prototype.clearOptionsMap = function() {
  this.getOptionsMap().clear();
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeSetResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeSetResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeSetResponse.displayName = 'proto.openstorage.api.VolumeSetResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeSetResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeSetResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeSetResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeSetResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    volume: (f = msg.getVolume()) && proto.openstorage.api.Volume.toObject(includeInstance, f),
    volumeResponse: (f = msg.getVolumeResponse()) && proto.openstorage.api.VolumeResponse.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeSetResponse}
 */
proto.openstorage.api.VolumeSetResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeSetResponse;
  return proto.openstorage.api.VolumeSetResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeSetResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeSetResponse}
 */
proto.openstorage.api.VolumeSetResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.Volume;
      reader.readMessage(value,proto.openstorage.api.Volume.deserializeBinaryFromReader);
      msg.setVolume(value);
      break;
    case 2:
      var value = new proto.openstorage.api.VolumeResponse;
      reader.readMessage(value,proto.openstorage.api.VolumeResponse.deserializeBinaryFromReader);
      msg.setVolumeResponse(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeSetResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeSetResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeSetResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeSetResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolume();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.Volume.serializeBinaryToWriter
    );
  }
  f = message.getVolumeResponse();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.openstorage.api.VolumeResponse.serializeBinaryToWriter
    );
  }
};


/**
 * optional Volume volume = 1;
 * @return {?proto.openstorage.api.Volume}
 */
proto.openstorage.api.VolumeSetResponse.prototype.getVolume = function() {
  return /** @type{?proto.openstorage.api.Volume} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.Volume, 1));
};


/** @param {?proto.openstorage.api.Volume|undefined} value */
proto.openstorage.api.VolumeSetResponse.prototype.setVolume = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.VolumeSetResponse.prototype.clearVolume = function() {
  this.setVolume(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeSetResponse.prototype.hasVolume = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional VolumeResponse volume_response = 2;
 * @return {?proto.openstorage.api.VolumeResponse}
 */
proto.openstorage.api.VolumeSetResponse.prototype.getVolumeResponse = function() {
  return /** @type{?proto.openstorage.api.VolumeResponse} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeResponse, 2));
};


/** @param {?proto.openstorage.api.VolumeResponse|undefined} value */
proto.openstorage.api.VolumeSetResponse.prototype.setVolumeResponse = function(value) {
  jspb.Message.setWrapperField(this, 2, value);
};


proto.openstorage.api.VolumeSetResponse.prototype.clearVolumeResponse = function() {
  this.setVolumeResponse(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeSetResponse.prototype.hasVolumeResponse = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.SnapCreateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.SnapCreateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.SnapCreateRequest.displayName = 'proto.openstorage.api.SnapCreateRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.SnapCreateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.SnapCreateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.SnapCreateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.SnapCreateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    locator: (f = msg.getLocator()) && proto.openstorage.api.VolumeLocator.toObject(includeInstance, f),
    readonly: jspb.Message.getFieldWithDefault(msg, 3, false)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.SnapCreateRequest}
 */
proto.openstorage.api.SnapCreateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.SnapCreateRequest;
  return proto.openstorage.api.SnapCreateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.SnapCreateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.SnapCreateRequest}
 */
proto.openstorage.api.SnapCreateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = new proto.openstorage.api.VolumeLocator;
      reader.readMessage(value,proto.openstorage.api.VolumeLocator.deserializeBinaryFromReader);
      msg.setLocator(value);
      break;
    case 3:
      var value = /** @type {boolean} */ (reader.readBool());
      msg.setReadonly(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.SnapCreateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.SnapCreateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.SnapCreateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.SnapCreateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getLocator();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.openstorage.api.VolumeLocator.serializeBinaryToWriter
    );
  }
  f = message.getReadonly();
  if (f) {
    writer.writeBool(
      3,
      f
    );
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.openstorage.api.SnapCreateRequest.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.SnapCreateRequest.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional VolumeLocator locator = 2;
 * @return {?proto.openstorage.api.VolumeLocator}
 */
proto.openstorage.api.SnapCreateRequest.prototype.getLocator = function() {
  return /** @type{?proto.openstorage.api.VolumeLocator} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeLocator, 2));
};


/** @param {?proto.openstorage.api.VolumeLocator|undefined} value */
proto.openstorage.api.SnapCreateRequest.prototype.setLocator = function(value) {
  jspb.Message.setWrapperField(this, 2, value);
};


proto.openstorage.api.SnapCreateRequest.prototype.clearLocator = function() {
  this.setLocator(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.SnapCreateRequest.prototype.hasLocator = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional bool readonly = 3;
 * Note that Boolean fields may be set to 0/1 when serialized from a Java server.
 * You should avoid comparisons like {@code val === true/false} in those cases.
 * @return {boolean}
 */
proto.openstorage.api.SnapCreateRequest.prototype.getReadonly = function() {
  return /** @type {boolean} */ (jspb.Message.getFieldWithDefault(this, 3, false));
};


/** @param {boolean} value */
proto.openstorage.api.SnapCreateRequest.prototype.setReadonly = function(value) {
  jspb.Message.setField(this, 3, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.SnapCreateResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.SnapCreateResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.SnapCreateResponse.displayName = 'proto.openstorage.api.SnapCreateResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.SnapCreateResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.SnapCreateResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.SnapCreateResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.SnapCreateResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    volumeCreateResponse: (f = msg.getVolumeCreateResponse()) && proto.openstorage.api.VolumeCreateResponse.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.SnapCreateResponse}
 */
proto.openstorage.api.SnapCreateResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.SnapCreateResponse;
  return proto.openstorage.api.SnapCreateResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.SnapCreateResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.SnapCreateResponse}
 */
proto.openstorage.api.SnapCreateResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.VolumeCreateResponse;
      reader.readMessage(value,proto.openstorage.api.VolumeCreateResponse.deserializeBinaryFromReader);
      msg.setVolumeCreateResponse(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.SnapCreateResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.SnapCreateResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.SnapCreateResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.SnapCreateResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolumeCreateResponse();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.VolumeCreateResponse.serializeBinaryToWriter
    );
  }
};


/**
 * optional VolumeCreateResponse volume_create_response = 1;
 * @return {?proto.openstorage.api.VolumeCreateResponse}
 */
proto.openstorage.api.SnapCreateResponse.prototype.getVolumeCreateResponse = function() {
  return /** @type{?proto.openstorage.api.VolumeCreateResponse} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeCreateResponse, 1));
};


/** @param {?proto.openstorage.api.VolumeCreateResponse|undefined} value */
proto.openstorage.api.SnapCreateResponse.prototype.setVolumeCreateResponse = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.SnapCreateResponse.prototype.clearVolumeCreateResponse = function() {
  this.setVolumeCreateResponse(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.SnapCreateResponse.prototype.hasVolumeCreateResponse = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeInfo = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeInfo, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeInfo.displayName = 'proto.openstorage.api.VolumeInfo';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeInfo.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeInfo.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeInfo} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeInfo.toObject = function(includeInstance, msg) {
  var f, obj = {
    volumeId: jspb.Message.getFieldWithDefault(msg, 1, ""),
    path: jspb.Message.getFieldWithDefault(msg, 2, ""),
    storage: (f = msg.getStorage()) && proto.openstorage.api.VolumeSpec.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeInfo}
 */
proto.openstorage.api.VolumeInfo.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeInfo;
  return proto.openstorage.api.VolumeInfo.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeInfo} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeInfo}
 */
proto.openstorage.api.VolumeInfo.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setVolumeId(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
      break;
    case 3:
      var value = new proto.openstorage.api.VolumeSpec;
      reader.readMessage(value,proto.openstorage.api.VolumeSpec.deserializeBinaryFromReader);
      msg.setStorage(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeInfo.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeInfo.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeInfo} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeInfo.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolumeId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getStorage();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.openstorage.api.VolumeSpec.serializeBinaryToWriter
    );
  }
};


/**
 * optional string volume_id = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeInfo.prototype.getVolumeId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeInfo.prototype.setVolumeId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional string path = 2;
 * @return {string}
 */
proto.openstorage.api.VolumeInfo.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeInfo.prototype.setPath = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional VolumeSpec storage = 3;
 * @return {?proto.openstorage.api.VolumeSpec}
 */
proto.openstorage.api.VolumeInfo.prototype.getStorage = function() {
  return /** @type{?proto.openstorage.api.VolumeSpec} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeSpec, 3));
};


/** @param {?proto.openstorage.api.VolumeSpec|undefined} value */
proto.openstorage.api.VolumeInfo.prototype.setStorage = function(value) {
  jspb.Message.setWrapperField(this, 3, value);
};


proto.openstorage.api.VolumeInfo.prototype.clearStorage = function() {
  this.setStorage(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeInfo.prototype.hasStorage = function() {
  return jspb.Message.getField(this, 3) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeConsumer = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeConsumer, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeConsumer.displayName = 'proto.openstorage.api.VolumeConsumer';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeConsumer.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeConsumer.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeConsumer} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeConsumer.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    namespace: jspb.Message.getFieldWithDefault(msg, 2, ""),
    type: jspb.Message.getFieldWithDefault(msg, 3, ""),
    nodeId: jspb.Message.getFieldWithDefault(msg, 4, ""),
    ownerName: jspb.Message.getFieldWithDefault(msg, 5, ""),
    ownerType: jspb.Message.getFieldWithDefault(msg, 6, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeConsumer}
 */
proto.openstorage.api.VolumeConsumer.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeConsumer;
  return proto.openstorage.api.VolumeConsumer.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeConsumer} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeConsumer}
 */
proto.openstorage.api.VolumeConsumer.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setNamespace(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setType(value);
      break;
    case 4:
      var value = /** @type {string} */ (reader.readString());
      msg.setNodeId(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setOwnerName(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setOwnerType(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeConsumer.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeConsumer.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeConsumer} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeConsumer.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getNamespace();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getType();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getNodeId();
  if (f.length > 0) {
    writer.writeString(
      4,
      f
    );
  }
  f = message.getOwnerName();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getOwnerType();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeConsumer.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeConsumer.prototype.setName = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional string namespace = 2;
 * @return {string}
 */
proto.openstorage.api.VolumeConsumer.prototype.getNamespace = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeConsumer.prototype.setNamespace = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional string type = 3;
 * @return {string}
 */
proto.openstorage.api.VolumeConsumer.prototype.getType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeConsumer.prototype.setType = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * optional string node_id = 4;
 * @return {string}
 */
proto.openstorage.api.VolumeConsumer.prototype.getNodeId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 4, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeConsumer.prototype.setNodeId = function(value) {
  jspb.Message.setField(this, 4, value);
};


/**
 * optional string owner_name = 5;
 * @return {string}
 */
proto.openstorage.api.VolumeConsumer.prototype.getOwnerName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeConsumer.prototype.setOwnerName = function(value) {
  jspb.Message.setField(this, 5, value);
};


/**
 * optional string owner_type = 6;
 * @return {string}
 */
proto.openstorage.api.VolumeConsumer.prototype.getOwnerType = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeConsumer.prototype.setOwnerType = function(value) {
  jspb.Message.setField(this, 6, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.GraphDriverChanges = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.GraphDriverChanges, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.GraphDriverChanges.displayName = 'proto.openstorage.api.GraphDriverChanges';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.GraphDriverChanges.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.GraphDriverChanges.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.GraphDriverChanges} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.GraphDriverChanges.toObject = function(includeInstance, msg) {
  var f, obj = {
    path: jspb.Message.getFieldWithDefault(msg, 1, ""),
    kind: jspb.Message.getFieldWithDefault(msg, 2, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.GraphDriverChanges}
 */
proto.openstorage.api.GraphDriverChanges.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.GraphDriverChanges;
  return proto.openstorage.api.GraphDriverChanges.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.GraphDriverChanges} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.GraphDriverChanges}
 */
proto.openstorage.api.GraphDriverChanges.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setPath(value);
      break;
    case 2:
      var value = /** @type {!proto.openstorage.api.GraphDriverChangeType} */ (reader.readEnum());
      msg.setKind(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.GraphDriverChanges.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.GraphDriverChanges.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.GraphDriverChanges} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.GraphDriverChanges.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getPath();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getKind();
  if (f !== 0.0) {
    writer.writeEnum(
      2,
      f
    );
  }
};


/**
 * optional string path = 1;
 * @return {string}
 */
proto.openstorage.api.GraphDriverChanges.prototype.getPath = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.GraphDriverChanges.prototype.setPath = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional GraphDriverChangeType kind = 2;
 * @return {!proto.openstorage.api.GraphDriverChangeType}
 */
proto.openstorage.api.GraphDriverChanges.prototype.getKind = function() {
  return /** @type {!proto.openstorage.api.GraphDriverChangeType} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/** @param {!proto.openstorage.api.GraphDriverChangeType} value */
proto.openstorage.api.GraphDriverChanges.prototype.setKind = function(value) {
  jspb.Message.setField(this, 2, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterResponse.displayName = 'proto.openstorage.api.ClusterResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    error: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterResponse}
 */
proto.openstorage.api.ClusterResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterResponse;
  return proto.openstorage.api.ClusterResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterResponse}
 */
proto.openstorage.api.ClusterResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string error = 1;
 * @return {string}
 */
proto.openstorage.api.ClusterResponse.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.ClusterResponse.prototype.setError = function(value) {
  jspb.Message.setField(this, 1, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ActiveRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ActiveRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ActiveRequest.displayName = 'proto.openstorage.api.ActiveRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ActiveRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ActiveRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ActiveRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ActiveRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    reqestkvMap: (f = msg.getReqestkvMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ActiveRequest}
 */
proto.openstorage.api.ActiveRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ActiveRequest;
  return proto.openstorage.api.ActiveRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ActiveRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ActiveRequest}
 */
proto.openstorage.api.ActiveRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getReqestkvMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readInt64, jspb.BinaryReader.prototype.readString);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ActiveRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ActiveRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ActiveRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ActiveRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getReqestkvMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeInt64, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * map<int64, string> ReqestKV = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<number,string>}
 */
proto.openstorage.api.ActiveRequest.prototype.getReqestkvMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<number,string>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      null));
};


proto.openstorage.api.ActiveRequest.prototype.clearReqestkvMap = function() {
  this.getReqestkvMap().clear();
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ActiveRequests = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.openstorage.api.ActiveRequests.repeatedFields_, null);
};
goog.inherits(proto.openstorage.api.ActiveRequests, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ActiveRequests.displayName = 'proto.openstorage.api.ActiveRequests';
}
/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.openstorage.api.ActiveRequests.repeatedFields_ = [2];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ActiveRequests.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ActiveRequests.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ActiveRequests} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ActiveRequests.toObject = function(includeInstance, msg) {
  var f, obj = {
    requestcount: jspb.Message.getFieldWithDefault(msg, 1, 0),
    activerequestList: jspb.Message.toObjectList(msg.getActiverequestList(),
    proto.openstorage.api.ActiveRequest.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ActiveRequests}
 */
proto.openstorage.api.ActiveRequests.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ActiveRequests;
  return proto.openstorage.api.ActiveRequests.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ActiveRequests} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ActiveRequests}
 */
proto.openstorage.api.ActiveRequests.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setRequestcount(value);
      break;
    case 2:
      var value = new proto.openstorage.api.ActiveRequest;
      reader.readMessage(value,proto.openstorage.api.ActiveRequest.deserializeBinaryFromReader);
      msg.addActiverequest(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ActiveRequests.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ActiveRequests.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ActiveRequests} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ActiveRequests.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getRequestcount();
  if (f !== 0) {
    writer.writeInt64(
      1,
      f
    );
  }
  f = message.getActiverequestList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      2,
      f,
      proto.openstorage.api.ActiveRequest.serializeBinaryToWriter
    );
  }
};


/**
 * optional int64 RequestCount = 1;
 * @return {number}
 */
proto.openstorage.api.ActiveRequests.prototype.getRequestcount = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {number} value */
proto.openstorage.api.ActiveRequests.prototype.setRequestcount = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * repeated ActiveRequest ActiveRequest = 2;
 * @return {!Array.<!proto.openstorage.api.ActiveRequest>}
 */
proto.openstorage.api.ActiveRequests.prototype.getActiverequestList = function() {
  return /** @type{!Array.<!proto.openstorage.api.ActiveRequest>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.openstorage.api.ActiveRequest, 2));
};


/** @param {!Array.<!proto.openstorage.api.ActiveRequest>} value */
proto.openstorage.api.ActiveRequests.prototype.setActiverequestList = function(value) {
  jspb.Message.setRepeatedWrapperField(this, 2, value);
};


/**
 * @param {!proto.openstorage.api.ActiveRequest=} opt_value
 * @param {number=} opt_index
 * @return {!proto.openstorage.api.ActiveRequest}
 */
proto.openstorage.api.ActiveRequests.prototype.addActiverequest = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 2, opt_value, proto.openstorage.api.ActiveRequest, opt_index);
};


proto.openstorage.api.ActiveRequests.prototype.clearActiverequestList = function() {
  this.setActiverequestList([]);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.GroupSnapCreateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.GroupSnapCreateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.GroupSnapCreateRequest.displayName = 'proto.openstorage.api.GroupSnapCreateRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.GroupSnapCreateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.GroupSnapCreateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.GroupSnapCreateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.GroupSnapCreateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    labelsMap: (f = msg.getLabelsMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.GroupSnapCreateRequest}
 */
proto.openstorage.api.GroupSnapCreateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.GroupSnapCreateRequest;
  return proto.openstorage.api.GroupSnapCreateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.GroupSnapCreateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.GroupSnapCreateRequest}
 */
proto.openstorage.api.GroupSnapCreateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = msg.getLabelsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.GroupSnapCreateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.GroupSnapCreateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.GroupSnapCreateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.GroupSnapCreateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getLabelsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(2, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.openstorage.api.GroupSnapCreateRequest.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.GroupSnapCreateRequest.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * map<string, string> Labels = 2;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.openstorage.api.GroupSnapCreateRequest.prototype.getLabelsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 2, opt_noLazyCreate,
      null));
};


proto.openstorage.api.GroupSnapCreateRequest.prototype.clearLabelsMap = function() {
  this.getLabelsMap().clear();
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.GroupSnapCreateResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.GroupSnapCreateResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.GroupSnapCreateResponse.displayName = 'proto.openstorage.api.GroupSnapCreateResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.GroupSnapCreateResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.GroupSnapCreateResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.GroupSnapCreateResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.GroupSnapCreateResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    snapshotsMap: (f = msg.getSnapshotsMap()) ? f.toObject(includeInstance, proto.openstorage.api.SnapCreateResponse.toObject) : [],
    error: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.GroupSnapCreateResponse}
 */
proto.openstorage.api.GroupSnapCreateResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.GroupSnapCreateResponse;
  return proto.openstorage.api.GroupSnapCreateResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.GroupSnapCreateResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.GroupSnapCreateResponse}
 */
proto.openstorage.api.GroupSnapCreateResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = msg.getSnapshotsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.openstorage.api.SnapCreateResponse.deserializeBinaryFromReader);
         });
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setError(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.GroupSnapCreateResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.GroupSnapCreateResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.GroupSnapCreateResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.GroupSnapCreateResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSnapshotsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(1, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.openstorage.api.SnapCreateResponse.serializeBinaryToWriter);
  }
  f = message.getError();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * map<string, SnapCreateResponse> snapshots = 1;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.openstorage.api.SnapCreateResponse>}
 */
proto.openstorage.api.GroupSnapCreateResponse.prototype.getSnapshotsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.openstorage.api.SnapCreateResponse>} */ (
      jspb.Message.getMapField(this, 1, opt_noLazyCreate,
      proto.openstorage.api.SnapCreateResponse));
};


proto.openstorage.api.GroupSnapCreateResponse.prototype.clearSnapshotsMap = function() {
  this.getSnapshotsMap().clear();
};


/**
 * optional string error = 2;
 * @return {string}
 */
proto.openstorage.api.GroupSnapCreateResponse.prototype.getError = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/** @param {string} value */
proto.openstorage.api.GroupSnapCreateResponse.prototype.setError = function(value) {
  jspb.Message.setField(this, 2, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.StorageNode = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.openstorage.api.StorageNode.repeatedFields_, null);
};
goog.inherits(proto.openstorage.api.StorageNode, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.StorageNode.displayName = 'proto.openstorage.api.StorageNode';
}
/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.openstorage.api.StorageNode.repeatedFields_ = [10];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.StorageNode.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.StorageNode.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.StorageNode} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.StorageNode.toObject = function(includeInstance, msg) {
  var f, obj = {
    id: jspb.Message.getFieldWithDefault(msg, 1, ""),
    cpu: +jspb.Message.getFieldWithDefault(msg, 2, 0.0),
    memTotal: jspb.Message.getFieldWithDefault(msg, 3, 0),
    memUsed: jspb.Message.getFieldWithDefault(msg, 4, 0),
    memFree: jspb.Message.getFieldWithDefault(msg, 5, 0),
    avgLoad: jspb.Message.getFieldWithDefault(msg, 6, 0),
    status: jspb.Message.getFieldWithDefault(msg, 7, 0),
    disksMap: (f = msg.getDisksMap()) ? f.toObject(includeInstance, proto.openstorage.api.StorageResource.toObject) : [],
    poolsList: jspb.Message.toObjectList(msg.getPoolsList(),
    proto.openstorage.api.StoragePool.toObject, includeInstance),
    mgmtIp: jspb.Message.getFieldWithDefault(msg, 11, ""),
    dataIp: jspb.Message.getFieldWithDefault(msg, 12, ""),
    hostname: jspb.Message.getFieldWithDefault(msg, 15, ""),
    nodeLabelsMap: (f = msg.getNodeLabelsMap()) ? f.toObject(includeInstance, undefined) : []
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.StorageNode}
 */
proto.openstorage.api.StorageNode.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.StorageNode;
  return proto.openstorage.api.StorageNode.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.StorageNode} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.StorageNode}
 */
proto.openstorage.api.StorageNode.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readDouble());
      msg.setCpu(value);
      break;
    case 3:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setMemTotal(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setMemUsed(value);
      break;
    case 5:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setMemFree(value);
      break;
    case 6:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setAvgLoad(value);
      break;
    case 7:
      var value = /** @type {!proto.openstorage.api.Status} */ (reader.readEnum());
      msg.setStatus(value);
      break;
    case 9:
      var value = msg.getDisksMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readMessage, proto.openstorage.api.StorageResource.deserializeBinaryFromReader);
         });
      break;
    case 10:
      var value = new proto.openstorage.api.StoragePool;
      reader.readMessage(value,proto.openstorage.api.StoragePool.deserializeBinaryFromReader);
      msg.addPools(value);
      break;
    case 11:
      var value = /** @type {string} */ (reader.readString());
      msg.setMgmtIp(value);
      break;
    case 12:
      var value = /** @type {string} */ (reader.readString());
      msg.setDataIp(value);
      break;
    case 15:
      var value = /** @type {string} */ (reader.readString());
      msg.setHostname(value);
      break;
    case 16:
      var value = msg.getNodeLabelsMap();
      reader.readMessage(value, function(message, reader) {
        jspb.Map.deserializeBinary(message, reader, jspb.BinaryReader.prototype.readString, jspb.BinaryReader.prototype.readString);
         });
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.StorageNode.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.StorageNode.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.StorageNode} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.StorageNode.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getCpu();
  if (f !== 0.0) {
    writer.writeDouble(
      2,
      f
    );
  }
  f = message.getMemTotal();
  if (f !== 0) {
    writer.writeUint64(
      3,
      f
    );
  }
  f = message.getMemUsed();
  if (f !== 0) {
    writer.writeUint64(
      4,
      f
    );
  }
  f = message.getMemFree();
  if (f !== 0) {
    writer.writeUint64(
      5,
      f
    );
  }
  f = message.getAvgLoad();
  if (f !== 0) {
    writer.writeInt64(
      6,
      f
    );
  }
  f = message.getStatus();
  if (f !== 0.0) {
    writer.writeEnum(
      7,
      f
    );
  }
  f = message.getDisksMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(9, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeMessage, proto.openstorage.api.StorageResource.serializeBinaryToWriter);
  }
  f = message.getPoolsList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      10,
      f,
      proto.openstorage.api.StoragePool.serializeBinaryToWriter
    );
  }
  f = message.getMgmtIp();
  if (f.length > 0) {
    writer.writeString(
      11,
      f
    );
  }
  f = message.getDataIp();
  if (f.length > 0) {
    writer.writeString(
      12,
      f
    );
  }
  f = message.getHostname();
  if (f.length > 0) {
    writer.writeString(
      15,
      f
    );
  }
  f = message.getNodeLabelsMap(true);
  if (f && f.getLength() > 0) {
    f.serializeBinary(16, writer, jspb.BinaryWriter.prototype.writeString, jspb.BinaryWriter.prototype.writeString);
  }
};


/**
 * optional string id = 1;
 * @return {string}
 */
proto.openstorage.api.StorageNode.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageNode.prototype.setId = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional double cpu = 2;
 * @return {number}
 */
proto.openstorage.api.StorageNode.prototype.getCpu = function() {
  return /** @type {number} */ (+jspb.Message.getFieldWithDefault(this, 2, 0.0));
};


/** @param {number} value */
proto.openstorage.api.StorageNode.prototype.setCpu = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional uint64 mem_total = 3;
 * @return {number}
 */
proto.openstorage.api.StorageNode.prototype.getMemTotal = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/** @param {number} value */
proto.openstorage.api.StorageNode.prototype.setMemTotal = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * optional uint64 mem_used = 4;
 * @return {number}
 */
proto.openstorage.api.StorageNode.prototype.getMemUsed = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/** @param {number} value */
proto.openstorage.api.StorageNode.prototype.setMemUsed = function(value) {
  jspb.Message.setField(this, 4, value);
};


/**
 * optional uint64 mem_free = 5;
 * @return {number}
 */
proto.openstorage.api.StorageNode.prototype.getMemFree = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 5, 0));
};


/** @param {number} value */
proto.openstorage.api.StorageNode.prototype.setMemFree = function(value) {
  jspb.Message.setField(this, 5, value);
};


/**
 * optional int64 avg_load = 6;
 * @return {number}
 */
proto.openstorage.api.StorageNode.prototype.getAvgLoad = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 6, 0));
};


/** @param {number} value */
proto.openstorage.api.StorageNode.prototype.setAvgLoad = function(value) {
  jspb.Message.setField(this, 6, value);
};


/**
 * optional Status status = 7;
 * @return {!proto.openstorage.api.Status}
 */
proto.openstorage.api.StorageNode.prototype.getStatus = function() {
  return /** @type {!proto.openstorage.api.Status} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/** @param {!proto.openstorage.api.Status} value */
proto.openstorage.api.StorageNode.prototype.setStatus = function(value) {
  jspb.Message.setField(this, 7, value);
};


/**
 * map<string, StorageResource> disks = 9;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,!proto.openstorage.api.StorageResource>}
 */
proto.openstorage.api.StorageNode.prototype.getDisksMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,!proto.openstorage.api.StorageResource>} */ (
      jspb.Message.getMapField(this, 9, opt_noLazyCreate,
      proto.openstorage.api.StorageResource));
};


proto.openstorage.api.StorageNode.prototype.clearDisksMap = function() {
  this.getDisksMap().clear();
};


/**
 * repeated StoragePool pools = 10;
 * @return {!Array.<!proto.openstorage.api.StoragePool>}
 */
proto.openstorage.api.StorageNode.prototype.getPoolsList = function() {
  return /** @type{!Array.<!proto.openstorage.api.StoragePool>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.openstorage.api.StoragePool, 10));
};


/** @param {!Array.<!proto.openstorage.api.StoragePool>} value */
proto.openstorage.api.StorageNode.prototype.setPoolsList = function(value) {
  jspb.Message.setRepeatedWrapperField(this, 10, value);
};


/**
 * @param {!proto.openstorage.api.StoragePool=} opt_value
 * @param {number=} opt_index
 * @return {!proto.openstorage.api.StoragePool}
 */
proto.openstorage.api.StorageNode.prototype.addPools = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 10, opt_value, proto.openstorage.api.StoragePool, opt_index);
};


proto.openstorage.api.StorageNode.prototype.clearPoolsList = function() {
  this.setPoolsList([]);
};


/**
 * optional string mgmt_ip = 11;
 * @return {string}
 */
proto.openstorage.api.StorageNode.prototype.getMgmtIp = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 11, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageNode.prototype.setMgmtIp = function(value) {
  jspb.Message.setField(this, 11, value);
};


/**
 * optional string data_ip = 12;
 * @return {string}
 */
proto.openstorage.api.StorageNode.prototype.getDataIp = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 12, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageNode.prototype.setDataIp = function(value) {
  jspb.Message.setField(this, 12, value);
};


/**
 * optional string hostname = 15;
 * @return {string}
 */
proto.openstorage.api.StorageNode.prototype.getHostname = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 15, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageNode.prototype.setHostname = function(value) {
  jspb.Message.setField(this, 15, value);
};


/**
 * map<string, string> node_labels = 16;
 * @param {boolean=} opt_noLazyCreate Do not create the map if
 * empty, instead returning `undefined`
 * @return {!jspb.Map<string,string>}
 */
proto.openstorage.api.StorageNode.prototype.getNodeLabelsMap = function(opt_noLazyCreate) {
  return /** @type {!jspb.Map<string,string>} */ (
      jspb.Message.getMapField(this, 16, opt_noLazyCreate,
      null));
};


proto.openstorage.api.StorageNode.prototype.clearNodeLabelsMap = function() {
  this.getNodeLabelsMap().clear();
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.StorageCluster = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.openstorage.api.StorageCluster.repeatedFields_, null);
};
goog.inherits(proto.openstorage.api.StorageCluster, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.StorageCluster.displayName = 'proto.openstorage.api.StorageCluster';
}
/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.openstorage.api.StorageCluster.repeatedFields_ = [4];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.StorageCluster.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.StorageCluster.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.StorageCluster} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.StorageCluster.toObject = function(includeInstance, msg) {
  var f, obj = {
    status: jspb.Message.getFieldWithDefault(msg, 1, 0),
    id: jspb.Message.getFieldWithDefault(msg, 2, ""),
    nodeId: jspb.Message.getFieldWithDefault(msg, 3, ""),
    nodesList: jspb.Message.toObjectList(msg.getNodesList(),
    proto.openstorage.api.StorageNode.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.StorageCluster}
 */
proto.openstorage.api.StorageCluster.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.StorageCluster;
  return proto.openstorage.api.StorageCluster.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.StorageCluster} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.StorageCluster}
 */
proto.openstorage.api.StorageCluster.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!proto.openstorage.api.Status} */ (reader.readEnum());
      msg.setStatus(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setId(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setNodeId(value);
      break;
    case 4:
      var value = new proto.openstorage.api.StorageNode;
      reader.readMessage(value,proto.openstorage.api.StorageNode.deserializeBinaryFromReader);
      msg.addNodes(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.StorageCluster.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.StorageCluster.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.StorageCluster} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.StorageCluster.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getStatus();
  if (f !== 0.0) {
    writer.writeEnum(
      1,
      f
    );
  }
  f = message.getId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getNodeId();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getNodesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      4,
      f,
      proto.openstorage.api.StorageNode.serializeBinaryToWriter
    );
  }
};


/**
 * optional Status status = 1;
 * @return {!proto.openstorage.api.Status}
 */
proto.openstorage.api.StorageCluster.prototype.getStatus = function() {
  return /** @type {!proto.openstorage.api.Status} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {!proto.openstorage.api.Status} value */
proto.openstorage.api.StorageCluster.prototype.setStatus = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional string id = 2;
 * @return {string}
 */
proto.openstorage.api.StorageCluster.prototype.getId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageCluster.prototype.setId = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional string node_id = 3;
 * @return {string}
 */
proto.openstorage.api.StorageCluster.prototype.getNodeId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/** @param {string} value */
proto.openstorage.api.StorageCluster.prototype.setNodeId = function(value) {
  jspb.Message.setField(this, 3, value);
};


/**
 * repeated StorageNode nodes = 4;
 * @return {!Array.<!proto.openstorage.api.StorageNode>}
 */
proto.openstorage.api.StorageCluster.prototype.getNodesList = function() {
  return /** @type{!Array.<!proto.openstorage.api.StorageNode>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.openstorage.api.StorageNode, 4));
};


/** @param {!Array.<!proto.openstorage.api.StorageNode>} value */
proto.openstorage.api.StorageCluster.prototype.setNodesList = function(value) {
  jspb.Message.setRepeatedWrapperField(this, 4, value);
};


/**
 * @param {!proto.openstorage.api.StorageNode=} opt_value
 * @param {number=} opt_index
 * @return {!proto.openstorage.api.StorageNode}
 */
proto.openstorage.api.StorageCluster.prototype.addNodes = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 4, opt_value, proto.openstorage.api.StorageNode, opt_index);
};


proto.openstorage.api.StorageCluster.prototype.clearNodesList = function() {
  this.setNodesList([]);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.OpenStorageVolumeCreateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.OpenStorageVolumeCreateRequest.displayName = 'proto.openstorage.api.OpenStorageVolumeCreateRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.OpenStorageVolumeCreateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.OpenStorageVolumeCreateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    spec: (f = msg.getSpec()) && proto.openstorage.api.VolumeSpec.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.OpenStorageVolumeCreateRequest}
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.OpenStorageVolumeCreateRequest;
  return proto.openstorage.api.OpenStorageVolumeCreateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.OpenStorageVolumeCreateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.OpenStorageVolumeCreateRequest}
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 2:
      var value = new proto.openstorage.api.VolumeSpec;
      reader.readMessage(value,proto.openstorage.api.VolumeSpec.deserializeBinaryFromReader);
      msg.setSpec(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.OpenStorageVolumeCreateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.OpenStorageVolumeCreateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getSpec();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      proto.openstorage.api.VolumeSpec.serializeBinaryToWriter
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.OpenStorageVolumeCreateRequest.prototype.setName = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional VolumeSpec spec = 2;
 * @return {?proto.openstorage.api.VolumeSpec}
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.prototype.getSpec = function() {
  return /** @type{?proto.openstorage.api.VolumeSpec} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeSpec, 2));
};


/** @param {?proto.openstorage.api.VolumeSpec|undefined} value */
proto.openstorage.api.OpenStorageVolumeCreateRequest.prototype.setSpec = function(value) {
  jspb.Message.setWrapperField(this, 2, value);
};


proto.openstorage.api.OpenStorageVolumeCreateRequest.prototype.clearSpec = function() {
  this.setSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.OpenStorageVolumeCreateRequest.prototype.hasSpec = function() {
  return jspb.Message.getField(this, 2) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.OpenStorageVolumeCreateResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.OpenStorageVolumeCreateResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.OpenStorageVolumeCreateResponse.displayName = 'proto.openstorage.api.OpenStorageVolumeCreateResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.OpenStorageVolumeCreateResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.OpenStorageVolumeCreateResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.OpenStorageVolumeCreateResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.OpenStorageVolumeCreateResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    volumeId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.OpenStorageVolumeCreateResponse}
 */
proto.openstorage.api.OpenStorageVolumeCreateResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.OpenStorageVolumeCreateResponse;
  return proto.openstorage.api.OpenStorageVolumeCreateResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.OpenStorageVolumeCreateResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.OpenStorageVolumeCreateResponse}
 */
proto.openstorage.api.OpenStorageVolumeCreateResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setVolumeId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.OpenStorageVolumeCreateResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.OpenStorageVolumeCreateResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.OpenStorageVolumeCreateResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.OpenStorageVolumeCreateResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolumeId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string volume_id = 1;
 * @return {string}
 */
proto.openstorage.api.OpenStorageVolumeCreateResponse.prototype.getVolumeId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.OpenStorageVolumeCreateResponse.prototype.setVolumeId = function(value) {
  jspb.Message.setField(this, 1, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeCreateFromVolumeIDRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeCreateFromVolumeIDRequest.displayName = 'proto.openstorage.api.VolumeCreateFromVolumeIDRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeCreateFromVolumeIDRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeCreateFromVolumeIDRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    name: jspb.Message.getFieldWithDefault(msg, 1, ""),
    parentId: jspb.Message.getFieldWithDefault(msg, 2, ""),
    spec: (f = msg.getSpec()) && proto.openstorage.api.VolumeSpec.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeCreateFromVolumeIDRequest}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeCreateFromVolumeIDRequest;
  return proto.openstorage.api.VolumeCreateFromVolumeIDRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeCreateFromVolumeIDRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeCreateFromVolumeIDRequest}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setName(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setParentId(value);
      break;
    case 3:
      var value = new proto.openstorage.api.VolumeSpec;
      reader.readMessage(value,proto.openstorage.api.VolumeSpec.deserializeBinaryFromReader);
      msg.setSpec(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeCreateFromVolumeIDRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeCreateFromVolumeIDRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getName();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getParentId();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getSpec();
  if (f != null) {
    writer.writeMessage(
      3,
      f,
      proto.openstorage.api.VolumeSpec.serializeBinaryToWriter
    );
  }
};


/**
 * optional string name = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.getName = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.setName = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional string parent_id = 2;
 * @return {string}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.getParentId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.setParentId = function(value) {
  jspb.Message.setField(this, 2, value);
};


/**
 * optional VolumeSpec spec = 3;
 * @return {?proto.openstorage.api.VolumeSpec}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.getSpec = function() {
  return /** @type{?proto.openstorage.api.VolumeSpec} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeSpec, 3));
};


/** @param {?proto.openstorage.api.VolumeSpec|undefined} value */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.setSpec = function(value) {
  jspb.Message.setWrapperField(this, 3, value);
};


proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.clearSpec = function() {
  this.setSpec(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDRequest.prototype.hasSpec = function() {
  return jspb.Message.getField(this, 3) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeCreateFromVolumeIDResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeCreateFromVolumeIDResponse.displayName = 'proto.openstorage.api.VolumeCreateFromVolumeIDResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeCreateFromVolumeIDResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeCreateFromVolumeIDResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    volumeId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeCreateFromVolumeIDResponse}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeCreateFromVolumeIDResponse;
  return proto.openstorage.api.VolumeCreateFromVolumeIDResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeCreateFromVolumeIDResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeCreateFromVolumeIDResponse}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setVolumeId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeCreateFromVolumeIDResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeCreateFromVolumeIDResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolumeId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string volume_id = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse.prototype.getVolumeId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeCreateFromVolumeIDResponse.prototype.setVolumeId = function(value) {
  jspb.Message.setField(this, 1, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeDeleteRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeDeleteRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeDeleteRequest.displayName = 'proto.openstorage.api.VolumeDeleteRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeDeleteRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeDeleteRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeDeleteRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeDeleteRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    volumeId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeDeleteRequest}
 */
proto.openstorage.api.VolumeDeleteRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeDeleteRequest;
  return proto.openstorage.api.VolumeDeleteRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeDeleteRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeDeleteRequest}
 */
proto.openstorage.api.VolumeDeleteRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setVolumeId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeDeleteRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeDeleteRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeDeleteRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeDeleteRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolumeId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string volume_id = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeDeleteRequest.prototype.getVolumeId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeDeleteRequest.prototype.setVolumeId = function(value) {
  jspb.Message.setField(this, 1, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeDeleteResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeDeleteResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeDeleteResponse.displayName = 'proto.openstorage.api.VolumeDeleteResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeDeleteResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeDeleteResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeDeleteResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeDeleteResponse.toObject = function(includeInstance, msg) {
  var f, obj = {

  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeDeleteResponse}
 */
proto.openstorage.api.VolumeDeleteResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeDeleteResponse;
  return proto.openstorage.api.VolumeDeleteResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeDeleteResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeDeleteResponse}
 */
proto.openstorage.api.VolumeDeleteResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeDeleteResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeDeleteResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeDeleteResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeDeleteResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeInspectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeInspectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeInspectRequest.displayName = 'proto.openstorage.api.VolumeInspectRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeInspectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeInspectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeInspectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeInspectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    volumeId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeInspectRequest}
 */
proto.openstorage.api.VolumeInspectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeInspectRequest;
  return proto.openstorage.api.VolumeInspectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeInspectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeInspectRequest}
 */
proto.openstorage.api.VolumeInspectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setVolumeId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeInspectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeInspectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeInspectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeInspectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolumeId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string volume_id = 1;
 * @return {string}
 */
proto.openstorage.api.VolumeInspectRequest.prototype.getVolumeId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.VolumeInspectRequest.prototype.setVolumeId = function(value) {
  jspb.Message.setField(this, 1, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeInspectResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeInspectResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeInspectResponse.displayName = 'proto.openstorage.api.VolumeInspectResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeInspectResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeInspectResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeInspectResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeInspectResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    volume: (f = msg.getVolume()) && proto.openstorage.api.Volume.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeInspectResponse}
 */
proto.openstorage.api.VolumeInspectResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeInspectResponse;
  return proto.openstorage.api.VolumeInspectResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeInspectResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeInspectResponse}
 */
proto.openstorage.api.VolumeInspectResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.Volume;
      reader.readMessage(value,proto.openstorage.api.Volume.deserializeBinaryFromReader);
      msg.setVolume(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeInspectResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeInspectResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeInspectResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeInspectResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolume();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.Volume.serializeBinaryToWriter
    );
  }
};


/**
 * optional Volume volume = 1;
 * @return {?proto.openstorage.api.Volume}
 */
proto.openstorage.api.VolumeInspectResponse.prototype.getVolume = function() {
  return /** @type{?proto.openstorage.api.Volume} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.Volume, 1));
};


/** @param {?proto.openstorage.api.Volume|undefined} value */
proto.openstorage.api.VolumeInspectResponse.prototype.setVolume = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.VolumeInspectResponse.prototype.clearVolume = function() {
  this.setVolume(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeInspectResponse.prototype.hasVolume = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeEnumerateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.VolumeEnumerateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeEnumerateRequest.displayName = 'proto.openstorage.api.VolumeEnumerateRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeEnumerateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeEnumerateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeEnumerateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeEnumerateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    locator: (f = msg.getLocator()) && proto.openstorage.api.VolumeLocator.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeEnumerateRequest}
 */
proto.openstorage.api.VolumeEnumerateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeEnumerateRequest;
  return proto.openstorage.api.VolumeEnumerateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeEnumerateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeEnumerateRequest}
 */
proto.openstorage.api.VolumeEnumerateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.VolumeLocator;
      reader.readMessage(value,proto.openstorage.api.VolumeLocator.deserializeBinaryFromReader);
      msg.setLocator(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeEnumerateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeEnumerateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeEnumerateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeEnumerateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getLocator();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.VolumeLocator.serializeBinaryToWriter
    );
  }
};


/**
 * optional VolumeLocator locator = 1;
 * @return {?proto.openstorage.api.VolumeLocator}
 */
proto.openstorage.api.VolumeEnumerateRequest.prototype.getLocator = function() {
  return /** @type{?proto.openstorage.api.VolumeLocator} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.VolumeLocator, 1));
};


/** @param {?proto.openstorage.api.VolumeLocator|undefined} value */
proto.openstorage.api.VolumeEnumerateRequest.prototype.setLocator = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.VolumeEnumerateRequest.prototype.clearLocator = function() {
  this.setLocator(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.VolumeEnumerateRequest.prototype.hasLocator = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.VolumeEnumerateResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, proto.openstorage.api.VolumeEnumerateResponse.repeatedFields_, null);
};
goog.inherits(proto.openstorage.api.VolumeEnumerateResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.VolumeEnumerateResponse.displayName = 'proto.openstorage.api.VolumeEnumerateResponse';
}
/**
 * List of repeated fields within this message type.
 * @private {!Array<number>}
 * @const
 */
proto.openstorage.api.VolumeEnumerateResponse.repeatedFields_ = [1];



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.VolumeEnumerateResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.VolumeEnumerateResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.VolumeEnumerateResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeEnumerateResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    volumesList: jspb.Message.toObjectList(msg.getVolumesList(),
    proto.openstorage.api.Volume.toObject, includeInstance)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.VolumeEnumerateResponse}
 */
proto.openstorage.api.VolumeEnumerateResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.VolumeEnumerateResponse;
  return proto.openstorage.api.VolumeEnumerateResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.VolumeEnumerateResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.VolumeEnumerateResponse}
 */
proto.openstorage.api.VolumeEnumerateResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.Volume;
      reader.readMessage(value,proto.openstorage.api.Volume.deserializeBinaryFromReader);
      msg.addVolumes(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.VolumeEnumerateResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.VolumeEnumerateResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.VolumeEnumerateResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.VolumeEnumerateResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getVolumesList();
  if (f.length > 0) {
    writer.writeRepeatedMessage(
      1,
      f,
      proto.openstorage.api.Volume.serializeBinaryToWriter
    );
  }
};


/**
 * repeated Volume volumes = 1;
 * @return {!Array.<!proto.openstorage.api.Volume>}
 */
proto.openstorage.api.VolumeEnumerateResponse.prototype.getVolumesList = function() {
  return /** @type{!Array.<!proto.openstorage.api.Volume>} */ (
    jspb.Message.getRepeatedWrapperField(this, proto.openstorage.api.Volume, 1));
};


/** @param {!Array.<!proto.openstorage.api.Volume>} value */
proto.openstorage.api.VolumeEnumerateResponse.prototype.setVolumesList = function(value) {
  jspb.Message.setRepeatedWrapperField(this, 1, value);
};


/**
 * @param {!proto.openstorage.api.Volume=} opt_value
 * @param {number=} opt_index
 * @return {!proto.openstorage.api.Volume}
 */
proto.openstorage.api.VolumeEnumerateResponse.prototype.addVolumes = function(opt_value, opt_index) {
  return jspb.Message.addToRepeatedWrapperField(this, 1, opt_value, proto.openstorage.api.Volume, opt_index);
};


proto.openstorage.api.VolumeEnumerateResponse.prototype.clearVolumesList = function() {
  this.setVolumesList([]);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterEnumerateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterEnumerateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterEnumerateRequest.displayName = 'proto.openstorage.api.ClusterEnumerateRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterEnumerateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterEnumerateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterEnumerateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterEnumerateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {

  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterEnumerateRequest}
 */
proto.openstorage.api.ClusterEnumerateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterEnumerateRequest;
  return proto.openstorage.api.ClusterEnumerateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterEnumerateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterEnumerateRequest}
 */
proto.openstorage.api.ClusterEnumerateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterEnumerateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterEnumerateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterEnumerateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterEnumerateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterEnumerateResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterEnumerateResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterEnumerateResponse.displayName = 'proto.openstorage.api.ClusterEnumerateResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterEnumerateResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterEnumerateResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterEnumerateResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterEnumerateResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    cluster: (f = msg.getCluster()) && proto.openstorage.api.StorageCluster.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterEnumerateResponse}
 */
proto.openstorage.api.ClusterEnumerateResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterEnumerateResponse;
  return proto.openstorage.api.ClusterEnumerateResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterEnumerateResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterEnumerateResponse}
 */
proto.openstorage.api.ClusterEnumerateResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.StorageCluster;
      reader.readMessage(value,proto.openstorage.api.StorageCluster.deserializeBinaryFromReader);
      msg.setCluster(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterEnumerateResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterEnumerateResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterEnumerateResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterEnumerateResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getCluster();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.StorageCluster.serializeBinaryToWriter
    );
  }
};


/**
 * optional StorageCluster cluster = 1;
 * @return {?proto.openstorage.api.StorageCluster}
 */
proto.openstorage.api.ClusterEnumerateResponse.prototype.getCluster = function() {
  return /** @type{?proto.openstorage.api.StorageCluster} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.StorageCluster, 1));
};


/** @param {?proto.openstorage.api.StorageCluster|undefined} value */
proto.openstorage.api.ClusterEnumerateResponse.prototype.setCluster = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.ClusterEnumerateResponse.prototype.clearCluster = function() {
  this.setCluster(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.ClusterEnumerateResponse.prototype.hasCluster = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterInspectRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterInspectRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterInspectRequest.displayName = 'proto.openstorage.api.ClusterInspectRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterInspectRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterInspectRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterInspectRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterInspectRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    nodeId: jspb.Message.getFieldWithDefault(msg, 1, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterInspectRequest}
 */
proto.openstorage.api.ClusterInspectRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterInspectRequest;
  return proto.openstorage.api.ClusterInspectRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterInspectRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterInspectRequest}
 */
proto.openstorage.api.ClusterInspectRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setNodeId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterInspectRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterInspectRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterInspectRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterInspectRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNodeId();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
};


/**
 * optional string node_id = 1;
 * @return {string}
 */
proto.openstorage.api.ClusterInspectRequest.prototype.getNodeId = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/** @param {string} value */
proto.openstorage.api.ClusterInspectRequest.prototype.setNodeId = function(value) {
  jspb.Message.setField(this, 1, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterInspectResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterInspectResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterInspectResponse.displayName = 'proto.openstorage.api.ClusterInspectResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterInspectResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterInspectResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterInspectResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterInspectResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    node: (f = msg.getNode()) && proto.openstorage.api.StorageNode.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterInspectResponse}
 */
proto.openstorage.api.ClusterInspectResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterInspectResponse;
  return proto.openstorage.api.ClusterInspectResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterInspectResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterInspectResponse}
 */
proto.openstorage.api.ClusterInspectResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.StorageNode;
      reader.readMessage(value,proto.openstorage.api.StorageNode.deserializeBinaryFromReader);
      msg.setNode(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterInspectResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterInspectResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterInspectResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterInspectResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getNode();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.StorageNode.serializeBinaryToWriter
    );
  }
};


/**
 * optional StorageNode node = 1;
 * @return {?proto.openstorage.api.StorageNode}
 */
proto.openstorage.api.ClusterInspectResponse.prototype.getNode = function() {
  return /** @type{?proto.openstorage.api.StorageNode} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.StorageNode, 1));
};


/** @param {?proto.openstorage.api.StorageNode|undefined} value */
proto.openstorage.api.ClusterInspectResponse.prototype.setNode = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.ClusterInspectResponse.prototype.clearNode = function() {
  this.setNode(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.ClusterInspectResponse.prototype.hasNode = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterAlertEnumerateRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterAlertEnumerateRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterAlertEnumerateRequest.displayName = 'proto.openstorage.api.ClusterAlertEnumerateRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterAlertEnumerateRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterAlertEnumerateRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    timeStart: (f = msg.getTimeStart()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    timeEnd: (f = msg.getTimeEnd()) && google_protobuf_timestamp_pb.Timestamp.toObject(includeInstance, f),
    resource: jspb.Message.getFieldWithDefault(msg, 3, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterAlertEnumerateRequest}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterAlertEnumerateRequest;
  return proto.openstorage.api.ClusterAlertEnumerateRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterAlertEnumerateRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterAlertEnumerateRequest}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setTimeStart(value);
      break;
    case 2:
      var value = new google_protobuf_timestamp_pb.Timestamp;
      reader.readMessage(value,google_protobuf_timestamp_pb.Timestamp.deserializeBinaryFromReader);
      msg.setTimeEnd(value);
      break;
    case 3:
      var value = /** @type {!proto.openstorage.api.ResourceType} */ (reader.readEnum());
      msg.setResource(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterAlertEnumerateRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterAlertEnumerateRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getTimeStart();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getTimeEnd();
  if (f != null) {
    writer.writeMessage(
      2,
      f,
      google_protobuf_timestamp_pb.Timestamp.serializeBinaryToWriter
    );
  }
  f = message.getResource();
  if (f !== 0.0) {
    writer.writeEnum(
      3,
      f
    );
  }
};


/**
 * optional google.protobuf.Timestamp time_start = 1;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.getTimeStart = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 1));
};


/** @param {?proto.google.protobuf.Timestamp|undefined} value */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.setTimeStart = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.clearTimeStart = function() {
  this.setTimeStart(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.hasTimeStart = function() {
  return jspb.Message.getField(this, 1) != null;
};


/**
 * optional google.protobuf.Timestamp time_end = 2;
 * @return {?proto.google.protobuf.Timestamp}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.getTimeEnd = function() {
  return /** @type{?proto.google.protobuf.Timestamp} */ (
    jspb.Message.getWrapperField(this, google_protobuf_timestamp_pb.Timestamp, 2));
};


/** @param {?proto.google.protobuf.Timestamp|undefined} value */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.setTimeEnd = function(value) {
  jspb.Message.setWrapperField(this, 2, value);
};


proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.clearTimeEnd = function() {
  this.setTimeEnd(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.hasTimeEnd = function() {
  return jspb.Message.getField(this, 2) != null;
};


/**
 * optional ResourceType resource = 3;
 * @return {!proto.openstorage.api.ResourceType}
 */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.getResource = function() {
  return /** @type {!proto.openstorage.api.ResourceType} */ (jspb.Message.getFieldWithDefault(this, 3, 0));
};


/** @param {!proto.openstorage.api.ResourceType} value */
proto.openstorage.api.ClusterAlertEnumerateRequest.prototype.setResource = function(value) {
  jspb.Message.setField(this, 3, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterAlertEnumerateResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterAlertEnumerateResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterAlertEnumerateResponse.displayName = 'proto.openstorage.api.ClusterAlertEnumerateResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterAlertEnumerateResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterAlertEnumerateResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterAlertEnumerateResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertEnumerateResponse.toObject = function(includeInstance, msg) {
  var f, obj = {
    alerts: (f = msg.getAlerts()) && proto.openstorage.api.Alerts.toObject(includeInstance, f)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterAlertEnumerateResponse}
 */
proto.openstorage.api.ClusterAlertEnumerateResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterAlertEnumerateResponse;
  return proto.openstorage.api.ClusterAlertEnumerateResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterAlertEnumerateResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterAlertEnumerateResponse}
 */
proto.openstorage.api.ClusterAlertEnumerateResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = new proto.openstorage.api.Alerts;
      reader.readMessage(value,proto.openstorage.api.Alerts.deserializeBinaryFromReader);
      msg.setAlerts(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterAlertEnumerateResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterAlertEnumerateResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterAlertEnumerateResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertEnumerateResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getAlerts();
  if (f != null) {
    writer.writeMessage(
      1,
      f,
      proto.openstorage.api.Alerts.serializeBinaryToWriter
    );
  }
};


/**
 * optional Alerts alerts = 1;
 * @return {?proto.openstorage.api.Alerts}
 */
proto.openstorage.api.ClusterAlertEnumerateResponse.prototype.getAlerts = function() {
  return /** @type{?proto.openstorage.api.Alerts} */ (
    jspb.Message.getWrapperField(this, proto.openstorage.api.Alerts, 1));
};


/** @param {?proto.openstorage.api.Alerts|undefined} value */
proto.openstorage.api.ClusterAlertEnumerateResponse.prototype.setAlerts = function(value) {
  jspb.Message.setWrapperField(this, 1, value);
};


proto.openstorage.api.ClusterAlertEnumerateResponse.prototype.clearAlerts = function() {
  this.setAlerts(undefined);
};


/**
 * Returns whether this field is set.
 * @return {!boolean}
 */
proto.openstorage.api.ClusterAlertEnumerateResponse.prototype.hasAlerts = function() {
  return jspb.Message.getField(this, 1) != null;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterAlertClearRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterAlertClearRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterAlertClearRequest.displayName = 'proto.openstorage.api.ClusterAlertClearRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterAlertClearRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterAlertClearRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterAlertClearRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertClearRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    resource: jspb.Message.getFieldWithDefault(msg, 1, 0),
    alertId: jspb.Message.getFieldWithDefault(msg, 2, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterAlertClearRequest}
 */
proto.openstorage.api.ClusterAlertClearRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterAlertClearRequest;
  return proto.openstorage.api.ClusterAlertClearRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterAlertClearRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterAlertClearRequest}
 */
proto.openstorage.api.ClusterAlertClearRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!proto.openstorage.api.ResourceType} */ (reader.readEnum());
      msg.setResource(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setAlertId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterAlertClearRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterAlertClearRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterAlertClearRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertClearRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getResource();
  if (f !== 0.0) {
    writer.writeEnum(
      1,
      f
    );
  }
  f = message.getAlertId();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
};


/**
 * optional ResourceType resource = 1;
 * @return {!proto.openstorage.api.ResourceType}
 */
proto.openstorage.api.ClusterAlertClearRequest.prototype.getResource = function() {
  return /** @type {!proto.openstorage.api.ResourceType} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {!proto.openstorage.api.ResourceType} value */
proto.openstorage.api.ClusterAlertClearRequest.prototype.setResource = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional int64 alert_id = 2;
 * @return {number}
 */
proto.openstorage.api.ClusterAlertClearRequest.prototype.getAlertId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/** @param {number} value */
proto.openstorage.api.ClusterAlertClearRequest.prototype.setAlertId = function(value) {
  jspb.Message.setField(this, 2, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterAlertClearResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterAlertClearResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterAlertClearResponse.displayName = 'proto.openstorage.api.ClusterAlertClearResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterAlertClearResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterAlertClearResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterAlertClearResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertClearResponse.toObject = function(includeInstance, msg) {
  var f, obj = {

  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterAlertClearResponse}
 */
proto.openstorage.api.ClusterAlertClearResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterAlertClearResponse;
  return proto.openstorage.api.ClusterAlertClearResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterAlertClearResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterAlertClearResponse}
 */
proto.openstorage.api.ClusterAlertClearResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterAlertClearResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterAlertClearResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterAlertClearResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertClearResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterAlertEraseRequest = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterAlertEraseRequest, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterAlertEraseRequest.displayName = 'proto.openstorage.api.ClusterAlertEraseRequest';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterAlertEraseRequest.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterAlertEraseRequest.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterAlertEraseRequest} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertEraseRequest.toObject = function(includeInstance, msg) {
  var f, obj = {
    resource: jspb.Message.getFieldWithDefault(msg, 1, 0),
    alertId: jspb.Message.getFieldWithDefault(msg, 2, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterAlertEraseRequest}
 */
proto.openstorage.api.ClusterAlertEraseRequest.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterAlertEraseRequest;
  return proto.openstorage.api.ClusterAlertEraseRequest.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterAlertEraseRequest} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterAlertEraseRequest}
 */
proto.openstorage.api.ClusterAlertEraseRequest.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {!proto.openstorage.api.ResourceType} */ (reader.readEnum());
      msg.setResource(value);
      break;
    case 2:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setAlertId(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterAlertEraseRequest.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterAlertEraseRequest.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterAlertEraseRequest} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertEraseRequest.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getResource();
  if (f !== 0.0) {
    writer.writeEnum(
      1,
      f
    );
  }
  f = message.getAlertId();
  if (f !== 0) {
    writer.writeInt64(
      2,
      f
    );
  }
};


/**
 * optional ResourceType resource = 1;
 * @return {!proto.openstorage.api.ResourceType}
 */
proto.openstorage.api.ClusterAlertEraseRequest.prototype.getResource = function() {
  return /** @type {!proto.openstorage.api.ResourceType} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {!proto.openstorage.api.ResourceType} value */
proto.openstorage.api.ClusterAlertEraseRequest.prototype.setResource = function(value) {
  jspb.Message.setField(this, 1, value);
};


/**
 * optional int64 alert_id = 2;
 * @return {number}
 */
proto.openstorage.api.ClusterAlertEraseRequest.prototype.getAlertId = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 2, 0));
};


/** @param {number} value */
proto.openstorage.api.ClusterAlertEraseRequest.prototype.setAlertId = function(value) {
  jspb.Message.setField(this, 2, value);
};



/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.openstorage.api.ClusterAlertEraseResponse = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.openstorage.api.ClusterAlertEraseResponse, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  proto.openstorage.api.ClusterAlertEraseResponse.displayName = 'proto.openstorage.api.ClusterAlertEraseResponse';
}


if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.openstorage.api.ClusterAlertEraseResponse.prototype.toObject = function(opt_includeInstance) {
  return proto.openstorage.api.ClusterAlertEraseResponse.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.openstorage.api.ClusterAlertEraseResponse} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertEraseResponse.toObject = function(includeInstance, msg) {
  var f, obj = {

  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.openstorage.api.ClusterAlertEraseResponse}
 */
proto.openstorage.api.ClusterAlertEraseResponse.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.openstorage.api.ClusterAlertEraseResponse;
  return proto.openstorage.api.ClusterAlertEraseResponse.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.openstorage.api.ClusterAlertEraseResponse} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.openstorage.api.ClusterAlertEraseResponse}
 */
proto.openstorage.api.ClusterAlertEraseResponse.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.openstorage.api.ClusterAlertEraseResponse.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.openstorage.api.ClusterAlertEraseResponse.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.openstorage.api.ClusterAlertEraseResponse} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.openstorage.api.ClusterAlertEraseResponse.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
};


/**
 * @enum {number}
 */
proto.openstorage.api.Status = {
  STATUS_NONE: 0,
  STATUS_INIT: 1,
  STATUS_OK: 2,
  STATUS_OFFLINE: 3,
  STATUS_ERROR: 4,
  STATUS_NOT_IN_QUORUM: 5,
  STATUS_DECOMMISSION: 6,
  STATUS_MAINTENANCE: 7,
  STATUS_STORAGE_DOWN: 8,
  STATUS_STORAGE_DEGRADED: 9,
  STATUS_NEEDS_REBOOT: 10,
  STATUS_STORAGE_REBALANCE: 11,
  STATUS_STORAGE_DRIVE_REPLACE: 12,
  STATUS_NOT_IN_QUORUM_NO_STORAGE: 13,
  STATUS_MAX: 14
};

/**
 * @enum {number}
 */
proto.openstorage.api.DriverType = {
  DRIVER_TYPE_NONE: 0,
  DRIVER_TYPE_FILE: 1,
  DRIVER_TYPE_BLOCK: 2,
  DRIVER_TYPE_OBJECT: 3,
  DRIVER_TYPE_CLUSTERED: 4,
  DRIVER_TYPE_GRAPH: 5
};

/**
 * @enum {number}
 */
proto.openstorage.api.FSType = {
  FS_TYPE_NONE: 0,
  FS_TYPE_BTRFS: 1,
  FS_TYPE_EXT4: 2,
  FS_TYPE_FUSE: 3,
  FS_TYPE_NFS: 4,
  FS_TYPE_VFS: 5,
  FS_TYPE_XFS: 6,
  FS_TYPE_ZFS: 7
};

/**
 * @enum {number}
 */
proto.openstorage.api.GraphDriverChangeType = {
  GRAPH_DRIVER_CHANGE_TYPE_NONE: 0,
  GRAPH_DRIVER_CHANGE_TYPE_MODIFIED: 1,
  GRAPH_DRIVER_CHANGE_TYPE_ADDED: 2,
  GRAPH_DRIVER_CHANGE_TYPE_DELETED: 3
};

/**
 * @enum {number}
 */
proto.openstorage.api.SeverityType = {
  SEVERITY_TYPE_NONE: 0,
  SEVERITY_TYPE_ALARM: 1,
  SEVERITY_TYPE_WARNING: 2,
  SEVERITY_TYPE_NOTIFY: 3
};

/**
 * @enum {number}
 */
proto.openstorage.api.ResourceType = {
  RESOURCE_TYPE_NONE: 0,
  RESOURCE_TYPE_VOLUME: 1,
  RESOURCE_TYPE_NODE: 2,
  RESOURCE_TYPE_CLUSTER: 3,
  RESOURCE_TYPE_DRIVE: 4
};

/**
 * @enum {number}
 */
proto.openstorage.api.AlertActionType = {
  ALERT_ACTION_TYPE_NONE: 0,
  ALERT_ACTION_TYPE_DELETE: 1,
  ALERT_ACTION_TYPE_CREATE: 2,
  ALERT_ACTION_TYPE_UPDATE: 3
};

/**
 * @enum {number}
 */
proto.openstorage.api.VolumeActionParam = {
  VOLUME_ACTION_PARAM_NONE: 0,
  VOLUME_ACTION_PARAM_OFF: 1,
  VOLUME_ACTION_PARAM_ON: 2
};

/**
 * @enum {number}
 */
proto.openstorage.api.CosType = {
  NONE: 0,
  LOW: 1,
  MEDIUM: 2,
  HIGH: 3
};

/**
 * @enum {number}
 */
proto.openstorage.api.IoProfile = {
  IO_PROFILE_SEQUENTIAL: 0,
  IO_PROFILE_RANDOM: 1,
  IO_PROFILE_DB: 2,
  IO_PROFILE_DB_REMOTE: 3,
  IO_PROFILE_CMS: 4
};

/**
 * @enum {number}
 */
proto.openstorage.api.VolumeState = {
  VOLUME_STATE_NONE: 0,
  VOLUME_STATE_PENDING: 1,
  VOLUME_STATE_AVAILABLE: 2,
  VOLUME_STATE_ATTACHED: 3,
  VOLUME_STATE_DETACHED: 4,
  VOLUME_STATE_DETATCHING: 5,
  VOLUME_STATE_ERROR: 6,
  VOLUME_STATE_DELETED: 7,
  VOLUME_STATE_TRY_DETACHING: 8,
  VOLUME_STATE_RESTORE: 9
};

/**
 * @enum {number}
 */
proto.openstorage.api.VolumeStatus = {
  VOLUME_STATUS_NONE: 0,
  VOLUME_STATUS_NOT_PRESENT: 1,
  VOLUME_STATUS_UP: 2,
  VOLUME_STATUS_DOWN: 3,
  VOLUME_STATUS_DEGRADED: 4
};

/**
 * @enum {number}
 */
proto.openstorage.api.StorageMedium = {
  STORAGE_MEDIUM_MAGNETIC: 0,
  STORAGE_MEDIUM_SSD: 1,
  STORAGE_MEDIUM_NVME: 2
};

/**
 * @enum {number}
 */
proto.openstorage.api.ClusterNotify = {
  CLUSTER_NOTIFY_DOWN: 0
};

/**
 * @enum {number}
 */
proto.openstorage.api.AttachState = {
  ATTACH_STATE_EXTERNAL: 0,
  ATTACH_STATE_INTERNAL: 1,
  ATTACH_STATE_INTERNAL_SWITCH: 2
};

/**
 * @enum {number}
 */
proto.openstorage.api.OperationFlags = {
  OP_FLAGS_UNKNOWN: 0,
  OP_FLAGS_NONE: 1,
  OP_FLAGS_DETACH_FORCE: 2
};

goog.object.extend(exports, proto.openstorage.api);
