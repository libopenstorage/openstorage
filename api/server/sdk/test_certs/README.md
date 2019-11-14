Instructions for generating these keys:

1. Run `openssl req -x509 -nodes -days 18250 -newkey rsa:2048 -keyout server-key.pem -out server-cert.pem`
2. When asked for `Common Name (e.g. server FQDN or YOUR name) []:`, make sure to enter `localhost`. This is because the
   gRPC test server listens on `localhost`. 
3. Run `chmod 400 server-key.pem`

