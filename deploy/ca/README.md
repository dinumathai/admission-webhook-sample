# Commands to generate certificate

## Generate a CA certificate
Edit the ca.conf for change in Root CA details.
```
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -days 100000 -out ca.crt -extensions v3_req  -extensions v3_ca -config ca.conf
```

## Create a server certificate.
Edit the server.conf if service-name(admission-webhook.) and namespace(webhook) is different(admission-webhook.webhook.svc).
```
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -config server.conf
```
## Sign the server certificate with the above CA
```
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 100000 -extensions v3_req -extfile server.conf
```

## Create configmap with server.key and server.crt
```
kubectl create configmap -n webhook admission-webhook-cert --from-file=/path/to/server/cert
```

## Generate value for "webhooks/clientConfig/caBundle" in webhook-admission-configuration.yaml
```
cat ca.crt | base64 | tr -d '\n'
```