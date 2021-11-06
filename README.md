# Kubernetes Admission webhook using golang in minikube

[admission-webhook-sample](https://github.com/dinumathai/admission-webhook-sample) is a sample Kubernetes mutating admission webhook project written in golang.

### What is Dynamic Admission webhook ?

An admission webhook is an HTTPS service that is called by Kubernetes api-server when it receives a request(CREATED/UPDATED/DELETED Kubernetes resource). The webhook is called prior to persistence of the object, but after the request is authenticated and authorized. The webhook response contain the information whether to allow the Kubernetes request to proceed further. Also may contain information on changes to be done on the Kubernetes request.

There are two type of Dynamic Admission webhook -
1. Validating admission webhook
1. Mutating admission webhook

### What is Mutating admission webhook ?

Mutating admission webhooks are invoked first, and can modify objects send to the Kubernetes API server. This is usually used to inject/set some values to the Kubernetes object.

### What is Validating admission webhook ?

After all object modifications are complete and after the incoming object is validated by the API server, validating admission webhooks are invoked and can accept/reject requests. This is usually used to enforce custom policies.

## When admission webhook comes into picture ?
![admission webhook flow](./doc/persistance-flow.png)

Once the request is authenticated and authorized all the mutating admission webhook will be called, which may change the incoming object. Then the schema validation in done. And finally all the validating admission webhook are called. If all the webhooks allows the request the object is persisted to DB.

## Use cases

1. Injecting Sidecar containers by looking at the annotation of the deployment/pod. [Isio](https://istio.io/) sidecar is an example.
1. Enforce policies on kubernetes objects. For example [Open policy agent](https://www.openpolicyagent.org/) is a tool that is used for validation of kubernetes objects and its is done using admission webhook.
1. Auditing on Kubernetes object. As all the http request can be configured to go through admission webhook, we can implement auditing using it.

## Mutating Webhook sample Request and response

Request Body Sample:
```
{"kind":"AdmissionReview","apiVersion":"admission.k8s.io/v1","request":{"uid":"c3ce937b-1a18-4e45-a0e8-adf5c962f1e4","kind":{"group":"","version":"v1","kind":"Pod"},"resource":{"group":"","version":"v1","resource":"pods"},"requestKind":{"group":"","version":"v1","kind":"Pod"},"requestResource":{"group":"","version":"v1","resource":"pods"},"name":"nginx-pod","namespace":"webhook","operation":"CREATE","userInfo":{"username":"minikube-user","groups":["system:masters","system:authenticated"]},"object":{"kind":"Pod","apiVersion":"v1","metadata":{"name":"nginx-pod","namespace":"webhook","creationTimestamp":null,"labels":{"app":"nginx"},"annotations":{"inject-init-container":"true","kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"annotations\":{\"inject-init-container\":\"true\"},\"labels\":{\"app\":\"nginx\"},\"name\":\"nginx-pod\",\"namespace\":\"webhook\"},\"spec\":{\"containers\":[{\"image\":\"nginx:1.14.2\",\"imagePullPolicy\":\"IfNotPresent\",\"name\":\"nginx\",\"ports\":[{\"containerPort\":80,\"protocol\":\"TCP\"}]}]}}\n"},"managedFields":[{"manager":"kubectl-client-side-apply","operation":"Update","apiVersion":"v1","time":"2021-11-04T08:20:33Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:inject-init-container":{},"f:kubectl.kubernetes.io/last-applied-configuration":{}},"f:labels":{".":{},"f:app":{}}},"f:spec":{"f:containers":{"k:{\"name\":\"nginx\"}":{".":{},"f:image":{},"f:imagePullPolicy":{},"f:name":{},"f:ports":{".":{},"k:{\"containerPort\":80,\"protocol\":\"TCP\"}":{".":{},"f:containerPort":{},"f:protocol":{}}},"f:resources":{},"f:terminationMessagePath":{},"f:terminationMessagePolicy":{}}},"f:dnsPolicy":{},"f:enableServiceLinks":{},"f:restartPolicy":{},"f:schedulerName":{},"f:securityContext":{},"f:terminationGracePeriodSeconds":{}}}}]},"spec":{"volumes":[{"name":"kube-api-access-bhn9b","projected":{"sources":[{"serviceAccountToken":{"expirationSeconds":3607,"path":"token"}},{"configMap":{"name":"kube-root-ca.crt","items":[{"key":"ca.crt","path":"ca.crt"}]}},{"downwardAPI":{"items":[{"path":"namespace","fieldRef":{"apiVersion":"v1","fieldPath":"metadata.namespace"}}]}}]}}],"containers":[{"name":"nginx","image":"nginx:1.14.2","ports":[{"containerPort":80,"protocol":"TCP"}],"resources":{},"volumeMounts":[{"name":"kube-api-access-bhn9b","readOnly":true,"mountPath":"/var/run/secrets/kubernetes.io/serviceaccount"}],"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","imagePullPolicy":"IfNotPresent"}],"restartPolicy":"Always","terminationGracePeriodSeconds":30,"dnsPolicy":"ClusterFirst","serviceAccountName":"default","serviceAccount":"default","securityContext":{},"schedulerName":"default-scheduler","tolerations":[{"key":"node.kubernetes.io/not-ready","operator":"Exists","effect":"NoExecute","tolerationSeconds":300},{"key":"node.kubernetes.io/unreachable","operator":"Exists","effect":"NoExecute","tolerationSeconds":300}],"priority":0,"enableServiceLinks":true,"preemptionPolicy":"PreemptLowerPriority"},"status":{}},"oldObject":null,"dryRun":false,"options":{"kind":"CreateOptions","apiVersion":"meta.k8s.io/v1","fieldManager":"kubectl-client-side-apply"}}}
```

Response Sample validating admission webhook
```
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "response": {
    "uid": "<value from request.uid>",
    "allowed": false
  }
}
```
Response Sample mutating admission webhook - The `patch` and `patchType` is optional.
```
{
  "kind": "AdmissionReview",
  "apiVersion": "admission.k8s.io/v1",
  "response": {
    "uid": "20bf8370-b5f4-4a3c-acc4-c7652dacd096",
    "allowed": true,
    "patch": "WwoJCXsKCQkgICJvcCI6ICJhZGQiLAoJCSAgInBhdGgiOiAiL3NwZWMvaW5pdENvbnRhaW5lcnMiLAoJCSAgInZhbHVlIjogWwoJCQl7CgkJCSAgImltYWdlIjogImJ1c3lib3g6MS4yOCIsCgkJCSAgIm5hbWUiOiAiaW5pdC1teXNlcnZpY2UiLAoJCQkgICJjb21tYW5kIjogWwoJCQkJInNoIiwKCQkJCSItYyIsCgkJCQkiZWNobyAtLS0tLS0tLS0tLS0tLS0tLS0tRXhlY3V0ZWRfVGhlX0luSXRfQ29udGFpbmVyLS0tLS0tLS0tLS0tLS0tLS0iCgkJCSAgXSwKCQkJICAicmVzb3VyY2VzIjogewoJCQkJCgkJCSAgfQoJCQl9CgkJICBdCgkJfQoJICBd",
    "patchType": "JSONPatch"
  }
}
```

Base64 decoded `patch`
```
[
    {
        "op": "add",
        "path": "/spec/initContainers",
        "value": [
        {
            "image": "busybox:1.28",
            "name": "init-myservice",
            "command": [
            "sh",
            "-c",
            "echo -------------------Executed_The_InIt_Container-----------------"
            ],
            "resources": {
            
            }
        }
        ]
    }
]
```
## Build and deploy in minikube

To get the webhooks up and running in minikube. First we have have generate certificates for webhooks, bring up the webhook and then configure the minikube to use it. And finally test it :-).

### Prerequisites
1. Basic understanding of Kubernetes.
1. Minikube running in local machine.
1. kubectl.
1. openssl(optional)

### Create the certificate
The certificate needed for webhook is available at [deploy/ca/](deploy/ca) folder. Certificates are generated under the assumption that the namespace is `webhook` and the `service` name is `admission-webhook`. If any change in namespace or service name [deploy/ca/server.conf](deploy/ca/server.conf) must be updated and certificates needs to be regenerated. Commands to generate the all certificate files are available at [deploy/ca/README.md](deploy/ca/README.md).

### Deploy in minikube

1. Start minikube. By default `ValidatingAdmissionWebhook` and `MutatingAdmissionWebhook` will be enabled.
1. The certificates are created for the K8S service `admission-webhook` inside namespace `webhook`. If the service name or namespace is different, please re-created certificates. Refer [deploy/ca/README.md](deploy/ca/README.md)
1. Created Namespace, Deployment, Service and MutatingAdmissionWebhook objects.
```
git clone git@github.com:dinumathai/admission-webhook-sample.git
cd admission-webhook-sample
kubectl create namespace webhook
kubectl create configmap -n webhook admission-webhook-cert --from-file=deploy/ca/
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/service.yaml
```
1. Makes sure that the webhook pod is up and running - `kubectl get pods -n webhook`. Once the webhook is up, create the webhook object - `kubectl apply -f deploy/webhook-admission-configuration.yaml`.
1. Test - Create a pod with required annotation using file deploy/test-nginx-pod.yaml(`kubectl apply -f deploy/test-nginx-pod.yaml`). Verify that init containers are injected by webhook - `k describe pod  nginx-pod -n webhook` and `k logs nginx-pod -n webhook -c init-myservice`.

## Developers Guide

### What does this sample MutatingAdmissionWebhook do ?
A MutatingAdmissionWebhook that injects an init container to pod if `inject-init-container` annotation is present in `pod`/`deployment`.

### Building and Run locally
```
git clone git@github.com:dinumathai/admission-webhook-sample.git
cd admission-webhook-sample
go build github.com/dinumathai/admission-webhook-sample

# TO RUN IN HTTPS mode. set the below variables
# export SSL_CRT_FILE_NAME=deploy/ca/server.crt
# export SSL_KEY_FILE_NAME=deploy/ca/server.key
./admission-webhook-sample -stderrthreshold=INFO -v=3
```
OR
```
go run main.go -stderrthreshold=INFO -v=3
```
## Reference
1. https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
1. https://github.com/kubernetes/kubernetes/tree/release-1.9/test/images/webhook
