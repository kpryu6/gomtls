# Multi Cluster mTLS

## 1. mTLS with SPIRE/SPIFFE in a Single Cluster(Cluster1)

### Introduction
Simple mutual TLS with client and server using SPIRE/SPIFFE

### Installation

1. Install spire server and spire agent
```shell
$ cd cluster1/spire
$ kubectl apply -f spire-all.yaml
# to make register entry automatically
$ kubectl apply -f cluster-spiffeid.yaml
```

```shell
$ kubectl get po -n spire
NAME                           READY   STATUS    RESTARTS   AGE
spire-agent-7vp8j              3/3     Running   0          4s
spire-server-fdc7d5f75-phk7l   2/2     Running   0          14s
```

<br>

2. Build server files

```shell
$ cd cluster1/server
# Run private docker registry
$ docker run -d -p 10000:5000 --name helloworld registry:2
$ docker ps
CONTAINER ID   IMAGE        COMMAND                  CREATED       STATUS       PORTS                                         NAMES
1e8ae9de7dd7   registry:2   "/entrypoint.sh /etc¡¦"   2 weeks ago   Up 2 weeks   0.0.0.0:10000->5000/tcp, :::10000->5000/tcp   helloworld
```

<br>

```shell
$ ./build.sh
$ kubectl apply -f helloworld-server.yaml
```

```shell
$ kubectl get svc
NAME                TYPE           CLUSTER-IP     EXTERNAL-IP       PORT(S)          AGE
helloworld-server   LoadBalancer   10.98.183.35   <pending>         8123:31631/TCP   2d19h
kubernetes          ClusterIP      10.96.0.1      <none>            443/TCP          10d
```

<br>

3. Get server's EXTERNAL-IP using metalLB

```shell
$ cd cluster1/metalLB
$ kubectl apply -f metalLB.yaml
# NOTICE : Use the IPAddressPool that matches the subnet for the network
$ kubectl apply -f my-network.yaml
```

```shell
$ kubectl get svc
NAME                TYPE           CLUSTER-IP     EXTERNAL-IP       PORT(S)          AGE
helloworld-server   LoadBalancer   10.98.183.35   **10.10.0.132**    8123:31631/TCP   2d19h
kubernetes          ClusterIP      10.96.0.1      <none>            443/TCP          10d
```

```shell
$ kubectl get po
NAME                                 READY   STATUS    RESTARTS   AGE
helloworld-server-65fdc68dcd-qzwgv   1/1     Running   0          8m44s
```

```shell
$ kubectl logs -f helloworld-server-65fdc68dcd-qzwgv
Server (:8123) starting up...
Server SVID:  spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server
Serving on [::]:8123
```

<br>

4. Build client files
```shell
$ cd cluster1/client
$ ./build.sh
# Use Server's EXTERNAL-IP in args
$ kubectl apply -f helloworld-client.yaml
```

```shell
$ kubectl get po
NAME                                 READY   STATUS    RESTARTS   AGE
helloworld-client-5c45d7798-b77vp    1/1     Running   0          3s
helloworld-server-65fdc68dcd-qzwgv   1/1     Running   0          12m
```

<br>

5. Confirm Register Entries
```shell
$ cd cluster1/script
$ ./show-entries.sh 
Found 2 entries
Entry ID         : 01ef6e8b-1d11-4e4f-8abb-2ad826b5e7e9
SPIFFE ID        : spiffe://cluster1.org/ns/default/sa/default/app/helloworld-client
Parent ID        : spiffe://cluster1.org/spire/agent/k8s_psat/cluster1-cluster/d9e67413-5390-4106-92c4-37efb025dd76
Revision         : 0
X509-SVID TTL    : default
JWT-SVID TTL     : default
Selector         : k8s:pod-uid:0e8d6db5-3a67-456e-94d2-0fac05d3d607

Entry ID         : 143071e3-5e58-456d-8859-fbb578b7a72b
SPIFFE ID        : spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server
Parent ID        : spiffe://cluster1.org/spire/agent/k8s_psat/cluster1-cluster/d9e67413-5390-4106-92c4-37efb025dd76
Revision         : 0
X509-SVID TTL    : default
JWT-SVID TTL     : default
Selector         : k8s:pod-uid:75e3288f-c271-48bb-a935-fc4278702203
```
<br>

6. mTLS communication between client and server with spiffeid

**server**
```shell
$ kubectl logs -f helloworld-server-65fdc68dcd-qzwgv
Server (:8123) starting up...
Server's spiffeID:  spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server
Serving on [::]:8123
Server gets BoanLab Client's SPIFFEID : "spiffe://cluster1.org/ns/default/sa/default/app/helloworld-client"
Server gets BoanLab Client's SPIFFEID : "spiffe://cluster1.org/ns/default/sa/default/app/helloworld-client"
...
```

**client**
```shell
$ kubectl logs -f helloworld-client-5c45d7798-b77vp
Client starting up... 10.10.0.132:8123
Will send request every 20s...
This is Server Reply : Hello BoanLab Client.
Client gets Server's SPIFFEID : "spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server"
This is Server Reply : Hello BoanLab Client.
Client gets Server's SPIFFEID : "spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server"
...
```

<br>

## 2. mTLS with SPIRE/SPIFFE in a Multi Cluster(Cluster1 & Cluster2)

### Federation

1. Copy Bundle in each cluster's spire server

- **Cluster1**

```shell
$ cd cluster1/script
$ ./spire-server.sh
----------Bundle Show----------
# This is Cluster1's bundle
{
    "keys": [
        {
            ...
```

- **Cluster2**

```shell
$ cd cluster2/script
$ ./spire-server.sh
----------Bundle Show----------
# This is Cluster2's bundle
{
    "keys": [
        {
            ...
```

<br>

2. Paste Bundle in ClusterFederatedTrustDomain's trustDomainBundle

- **Cluster1**

```shell
$ kubectl get svc -n spire
spire-server-bundle-endpoint               LoadBalancer   10.100.204.215   **10.10.0.131**  8443:32204/TCP   28h
```

- **Cluster2**

```shell
$ kubectl get svc -n spire
spire-server-bundle-endpoint               LoadBalancer   10.100.204.215   **10.10.0.141**  8443:32204/TCP   28h
```

<br>

- **Cluster1**

```shell
$ cd ../spire
$ vi cluster-federated-trust-domain.yaml
spec:
  trustDomain: cluster2.org
  # Cluster 2's spire-server-bundle-endpoint EXTERNAL-IP
  bundleEndpointURL: https://10.10.0.141:8443
  bundleEndpointProfile:
    type: https_spiffe
    endpointSPIFFEID: spiffe://cluster2.org/spire/server
  # Paste cluster 2's bundle
  trustDomainBundle: |-
      {
          "keys": [
              {
                  ...
$ kubectl apply -f cluster-federated-trust-domain.yaml
```

- **Cluster2**

```shell
$ cd ../spire
$ vi cluster-federated-trust-domain.yaml
spec:
  trustDomain: cluster1.org
  # Cluster 1's spire-server-bundle-endpoint EXTERNAL-IP
  bundleEndpointURL: https://10.10.0.131:8443
  bundleEndpointProfile:
    type: https_spiffe
    endpointSPIFFEID: spiffe://cluster1.org/spire/server
  # Paste cluster 1's bundle
  trustDomainBundle: |-
      {
          "keys": [
              {
                  ...
$ kubectl apply -f cluster-federated-trust-domain.yaml
```

<br>

3. Add FederatesWith in ClusterSPIFFEID CRD

- **Cluster1**
  
```shell
$ vi cluster-spiffeid.yaml
  podSelector:
    matchLabels:
      spiffe.io/spire-managed-identity: "true"
  # Add this
  federatesWith: ["cluster2.org"]
$ kubectl apply -f cluster-spiffeid.yaml
```

- **Cluster2**
  
```shell
$ vi cluster-spiffeid.yaml
  podSelector:
    matchLabels:
      spiffe.io/spire-managed-identity: "true"
  # Add this
  federatesWith: ["cluster1.org"]
$ kubectl apply -f cluster-spiffeid.yaml
```

<br>

4. Confirm Federation and Entries

- **Cluster1**
  
```shell
$ cd cluster1/script
$ ./spire-server.sh
----------Bundle Show----------
{
    "keys": [
        {
            ...
}
----------Bundle List----------
****************************************
* cluster2.org
****************************************
{
    "keys": [
        {
            ...
}
----------Federation List----------
Found 1 federation relationship

Trust domain              : cluster2.org
Bundle endpoint URL       : https://10.10.0.141:8443
Bundle endpoint profile   : https_spiffe
Endpoint SPIFFE ID        : spiffe://cluster2.org/spire/server
```

<br>

- **Cluster2**
  
```shell
$ cd cluster2/script
$ ./spire-server.sh
----------Bundle Show----------
{
    "keys": [
        {
            ...
}
----------Bundle List----------
****************************************
* cluster1.org
****************************************
{
    "keys": [
        {
            ...
}
----------Federation List----------
Found 1 federation relationship

Trust domain              : cluster1.org
Bundle endpoint URL       : https://10.10.0.131:8443
Bundle endpoint profile   : https_spiffe
Endpoint SPIFFE ID        : spiffe://cluster1.org/spire/server
```

<br>

5. Check spire-server-bundle-endpoint can accessible

- **Cluster1**
  
```shell
$ curl 10.10.0.141:8443
Client sent an HTTP request to an HTTPS server.
```

- **Cluster2**
  
```shell
$ curl 10.10.0.131:8443
Client sent an HTTP request to an HTTPS server.
```

<br>

6. Verify if the Federation has been completed & Share trust bundle between clusters

- **Cluster1**

```shell
$ kubectl logs -f <spire-server-name>
```
time="2024-05-06T05:31:42Z" level=debug msg="**Federation relationship created**" authorized_as=local authorized_via=transport method=BatchCreateFederationRelationship request_id=d3170253-f34a-48dd-ae88-afdd9362a133 service=trustdomain.v1.TrustDomain subsystem_name=api trust_domain_id=cluster2.org

time="2024-05-06T05:32:04Z" level=debug msg="Polling for bundle update" subsystem_name=bundle_client trust_domain=cluster2.org
time="2024-05-06T05:32:04Z" level=info msg="**Bundle refreshed**" subsystem_name=bundle_client trust_domain=cluster2.org
time="2024-05-06T05:32:04Z" level=debug msg="Scheduling next bundle refresh" at="2024-05-06T06:08:04Z" subsystem_name=bundle_client trust_domain=cluster2.org

<br>

- **Cluster2**
  
```shell
$ kubectl logs -f <spire-server-name>
```
time="2024-05-06T05:31:48Z" level=debug msg="**Federation relationship created**" authorized_as=local authorized_via=transport method=BatchCreateFederationRelationship request_id=35914f40-1feb-42b9-a1ca-e257272f1b1f service=trustdomain.v1.TrustDomain subsystem_name=api trust_domain_id=cluster1.org

time="2024-05-06T05:32:20Z" level=debug msg="Polling for bundle update" subsystem_name=bundle_client trust_domain=cluster1.org
time="2024-05-06T05:32:20Z" level=info msg="**Bundle refreshed**" subsystem_name=bundle_client trust_domain=cluster1.org
time="2024-05-06T05:32:20Z" level=debug msg="Scheduling next bundle refresh" at="2024-05-06T06:08:20Z" subsystem_name=bundle_client trust_domain=cluster1.org

<br>

7. Build client files in Cluster2
```shell
$ cd cluster2/client
$ ./build.sh
# Use Cluster1 Server's EXTERNAL-IP in args
$ kubectl apply -f helloworld-client.yaml
```

```shell
$ kubectl get po
NAME                                 READY   STATUS    RESTARTS   AGE
helloworld-client-0a45d2848-4i8sa    1/1     Running   0          3s
```

<br>

8. Confirm Register Entries

- **Cluster1**
  
```shell
$ cd cluster1/script
$ ./show-entries.sh 
Found 2 entries
Entry ID         : 01ef6e8b-1d11-4e4f-8abb-2ad826b5e7e9
SPIFFE ID        : spiffe://cluster1.org/ns/default/sa/default/app/helloworld-client
Parent ID        : spiffe://cluster1.org/spire/agent/k8s_psat/cluster1-cluster/d9e67413-5390-4106-92c4-37efb025dd76
Revision         : 0
X509-SVID TTL    : default
JWT-SVID TTL     : default
Selector         : k8s:pod-uid:0e8d6db5-3a67-456e-94d2-0fac05d3d607
FederatesWith    : cluster1.org

Entry ID         : 143071e3-5e58-456d-8859-fbb578b7a72b
SPIFFE ID        : spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server
Parent ID        : spiffe://cluster1.org/spire/agent/k8s_psat/cluster1-cluster/d9e67413-5390-4106-92c4-37efb025dd76
Revision         : 0
X509-SVID TTL    : default
JWT-SVID TTL     : default
Selector         : k8s:pod-uid:75e3288f-c271-48bb-a935-fc4278702203
FederatesWith    : cluster2.org
```

- **Cluster2**
  
```shell
$ cd cluster2/script
$ ./show-entries.sh
Found 1 entry
Entry ID         : a424fa86-e766-489c-b058-a8fc5cda99c6
SPIFFE ID        : spiffe://cluster2.org/ns/default/sa/default/app/helloworld-client
Parent ID        : spiffe://cluster2.org/spire/agent/k8s_psat/cluster2-cluster/1c0f79ea-6dd0-4a6b-a2c1-3852cded86af
Revision         : 0
X509-SVID TTL    : default
JWT-SVID TTL     : default
Selector         : k8s:pod-uid:8c05dcea-c9b1-4c19-9969-9b3980392ed8
FederatesWith    : cluster1.org
```

<br>

9. mTLS communication between two clients and server with spiffeid

**Cluster2 client**

```shell
$ kubectl logs -f helloworld-client-0a45d2848-4i8sa
Client starting up... 10.10.0.132:8123
Will send request every 20s...
This is Server Reply : Hello BoanLab Client.
Client gets Server's SPIFFEID : "spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server"
This is Server Reply : Hello BoanLab Client.
Client gets Server's SPIFFEID : "spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server"
...
```

**Cluster1 server**

```shell
$ kubectl logs -f helloworld-server-65fdc68dcd-qzwgv
Server (:8123) starting up...
Server's spiffeID:  spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server
Serving on [::]:8123
Server gets BoanLab Client's SPIFFEID : "spiffe://cluster2.org/ns/default/sa/default/app/helloworld-client"
Server gets BoanLab Client's SPIFFEID : "spiffe://cluster1.org/ns/default/sa/default/app/helloworld-client"
Server gets BoanLab Client's SPIFFEID : "spiffe://cluster2.org/ns/default/sa/default/app/helloworld-client"
Server gets BoanLab Client's SPIFFEID : "spiffe://cluster1.org/ns/default/sa/default/app/helloworld-client"
...
```

# mTLS with SPIRE/SPIFFE in Istio Service Mesh



