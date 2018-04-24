# Python gRPC Client Library

## Build

Type `make`.

Requires `virtualenv`. To install type: `pip install virtualenv`

## Example

```python
#
# Install:
#   virtualenv sdk
#   source sdk/bin/activate
#   pip install grpcio grpcio-tools
#
# Run:
#   sudo -E bash -c "source sdk/bin/activate && python client.py"
#
# More info: https://grpc.io/docs/quickstart/python.html
#
import grpc
import api_pb2
import api_pb2_grpc

# Setup connection
channel = grpc.insecure_channel('unix:/var/lib/osd/driver/nfs-sdk.sock')
client = api_pb2_grpc.OpenStorageClusterStub(channel)

# Get cluster information
en_resp = client.Enumerate(api_pb2.ClusterEnumerateRequest())
print en_resp

# Get node info
n_resp = client.Inspect(api_pb2.ClusterInspectRequest(node_id=en_resp.cluster.nodes[0].id))
print n_resp
```

