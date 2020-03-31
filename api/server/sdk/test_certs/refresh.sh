#!/bin/sh

chmod 644 server.key server.crt
openssl genrsa -out server.key 2048
openssl req -new -sha256 -key server.key -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=localhost" -out server.csr
openssl x509 -req -in server.csr -CA insecure_ca.crt -CAkey insecure_ca.key -CAcreateserial -out server.crt -days 500 -sha256
chmod 400 server.key server.crt

rm -f insecure_ca.srl


