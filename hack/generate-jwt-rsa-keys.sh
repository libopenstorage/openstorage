#!/bin/sh
# https://cloud.google.com/iot/docs/how-tos/credentials/keys
#
# Generate RSA256
openssl genrsa -out rsa_private.pem 2048
openssl rsa -in rsa_private.pem -pubout -out rsa_public.pem

