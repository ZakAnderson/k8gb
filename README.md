# Oh My GLB

## Project Health

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Actions Status](https://github.com/AbsaOSS/ohmyglb/workflows/build/badge.svg)](https://github.com/AbsaOSS/ohmyglb/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/AbsaOSS/ohmyglb)](https://goreportcard.com/report/github.com/AbsaOSS/ohmyglb)
[![Helm Publish](https://github.com/AbsaOSS/ohmyglb/workflows/Helm%20Publish/badge.svg)](https://github.com/AbsaOSS/ohmyglb/actions?query=workflow%3A%22Helm+Publish%22)
[![Docker Pulls](https://img.shields.io/docker/pulls/absaoss/ohmyglb)](https://hub.docker.com/r/absaoss/ohmyglb)
[![Gosec](https://github.com/AbsaOSS/ohmyglb/workflows/Gosec/badge.svg)](https://github.com/AbsaOSS/ohmyglb/actions?query=workflow%3AGosec)

A Global Service Load Balancing solution with a focus on having cloud native qualities and work natively in a Kubernetes context.

- [Motivation and Architecture](#motivation-and-architecture)
- [Installation and Configuration](#installation-and-configuration)
    - [Installation with Helm3](#installation-with-helm3)
        - [Add ohmyglb Helm repository](#add-ohmyglb-helm-repository)
    - [Local Playground Install](#local-playground-install)
        - [Environment prerequisites](#environment-prerequisites)
        - [Running project locally](#running-project-locally)
        - [Verify installation](#verify-installation)
        - [Run integration tests](#run-integration-tests)
        - [Cleaning](#cleaning)
- [Sample demo](#sample-demo)
    - [Round Robin](#round-robin)
- [Metrics](#metrics)
    - [General metrics](#general-metrics)
    - [Custom resource specific metrics](#custom-resource-specific-metrics)

## Motivation and Architecture

Please see the extended documentation [here](/docs/index.md)

## Installation and Configuration

### Installation with Helm3

#### Add ohmyglb Helm repository

```sh
$ helm repo add ohmyglb https://absaoss.github.io/ohmyglb/
$ helm repo update
$ helm install ohmyglb ohmyglb/ohmyglb
```

See [values.yaml](https://github.com/AbsaOSS/ohmyglb/blob/master/chart/ohmyglb/values.yaml)
for customization options.

### Local Playground Install

#### Environment prerequisites

 - [install **GO 1.14**](https://golang.org/dl/)
 
 - [install **GIT**](https://git-scm.com/downloads)
    
 - install **gnu-sed** if you don't have it
    - If you are on a Mac, install sed by Homebrew
    ```shell script
    brew install gnu-sed
    ```
   
 - [install **Docker**](https://docs.docker.com/get-docker/)
    - ensure you are able to push/pull from your docker registry
    - to run multiple clusters reserve 8GB of memory
    
      ![](docs/images/docker_settings.png)
      <div>
        <sup>above screenshot is for <strong>Docker for Mac</strong> and that options for other Docker distributions may vary</sup>
      </div>

 - [install **Kubectl**](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to operate clusters

 - [install **Helm3**](https://helm.sh/docs/intro/install/) to get charts

 - [install **kind**](https://kind.sigs.k8s.io/) as tool for running local Kubernetes clusters
    - follow https://kind.sigs.k8s.io/docs/user/quick-start/


#### Running project locally

To spin-up a local environment using two Kind clusters and deploy a test application to both clusters, execute the command below: 
```shell script
make deploy-full-local-setup 
```

#### Verify installation

If local setup runs well, check if clusters are correctly installed 

```shell script
kubectl cluster-info --context kind-test-gslb1 && kubectl cluster-info --context kind-test-gslb2
```

Check if Etcd cluster is healthy
```shell script
kubectl run --rm -i --tty --env="ETCDCTL_API=3" --env="ETCDCTL_ENDPOINTS=http://etcd-cluster-client:2379" --namespace ohmyglb etcd-test --image quay.io/coreos/etcd --restart=Never -- /bin/sh -c 'etcdctl  member list' 
```
as expected output you will see three started pods: `etcd-cluster`

```shell script
...
c3261c079f6990a7, started, etcd-cluster-5bcpvf6ngz, http://etcd-cluster-5bcpvf6ngz.etcd-cluster.ohmyglb.svc:2380, http://etcd-cluster-5bcpvf6ngz.etcd-cluster.ohmyglb.svc:2379
eb6ead15c2b92606, started, etcd-cluster-6d8pxtpklm, http://etcd-cluster-6d8pxtpklm.etcd-cluster.ohmyglb.svc:2380, http://etcd-cluster-6d8pxtpklm.etcd-cluster.ohmyglb.svc:2379
eed5a40bbfb6ee97, started, etcd-cluster-xsjmwdkdf8, http://etcd-cluster-xsjmwdkdf8.etcd-cluster.ohmyglb.svc:2380, http://etcd-cluster-xsjmwdkdf8.etcd-cluster.ohmyglb.svc:2379
...
```

Cluster [test-gslb1](deploy/kind/cluster.yaml) is exposing external DNS on default port `:5053` 
while [test-gslb2](deploy/kind/cluster2.yaml) on port `:5054`.
```shell script
dig @localhost localtargets.app3.cloud.example.com -p 5053 && dig -p 5054 @localhost localtargets.app3.cloud.example.com
```
As expected result you should see **six A records** divided between nodes of both clusters.
```shell script
...
...
;; ANSWER SECTION:
localtargets.app3.cloud.example.com. 30 IN A    172.17.0.2
localtargets.app3.cloud.example.com. 30 IN A    172.17.0.5
localtargets.app3.cloud.example.com. 30 IN A    172.17.0.3
...
...
localtargets.app3.cloud.example.com. 30 IN A    172.17.0.8
localtargets.app3.cloud.example.com. 30 IN A    172.17.0.6
localtargets.app3.cloud.example.com. 30 IN A    172.17.0.7
```
Both clusters have [podinfo](https://github.com/stefanprodan/podinfo) installed on the top. 
Run following command and check if you get two json responses.
```shell script
curl localhost:80 -H "Host:app3.cloud.example.com" && curl localhost:81 -H "Host:app3.cloud.example.com"
```

#### Run integration tests

There is wide range of scenarios which **GSLB** provides and all of them are covered within [tests](terratest).
To check whether everything is running properly execute [terratests](https://terratest.gruntwork.io/) :

```shell script
make terratest
```

#### Cleaning

Clean up your local development clusters with
```shell script
make destroy-full-local-setup
```


## Sample demo

### Round Robin

Both clusters have [podinfo](https://github.com/stefanprodan/podinfo) installed on the top where each 
cluster has adjusted a different region. In this demo we will hit podinfo by `curl app3.cloud.example.com` and depending 
on region will podinfo return **us** or **eu**. In current round robin implementation are ip addresses randomly picked. 
See [Gslb manifest with round robin strategy](/deploy/crds/ohmyglb.absa.oss_v1beta1_gslb_cr.yaml)

Run several times command below and watch `message` field.
```shell script
make test-round-robin
```
As expected result you should see podinfo message changing

```text
{
  "hostname": "frontend-podinfo-856bb46677-8p45m",
  ...
  "message": "us",
  ...
}
```
```text
{
  "hostname": "frontend-podinfo-856bb46677-8p45m",
  ...
  "message": "eu",
  ...
}
```
## Metrics

OhMyGLB generates [Prometheus][prometheus]-compatible metrics.
Metrics endpoints are exposed via `-metrics` service in operator namespace and can be scraped by 3rd party tools:

``` yaml
spec:
...
  ports:
  - name: http-metrics
    port: 8383
    protocol: TCP
    targetPort: 8383
  - name: cr-metrics
    port: 8686
    protocol: TCP
    targetPort: 8686
```

Metrics can be also automatically discovered and monitored by [Prometheus Operator][prometheus-operator] via automatically generated [ServiceMonitor][service-monitor] CRDs , in case if [Prometheus Operator][prometheus-operator]  is deployed into the cluster.

### General metrics

[controller-runtime][controller-runtime-metrics] standard metrics, extended with OhMyGLB operator-specific metrics listed below:

#### `healthy_records_total`

Number of healthy records observed by OhMyGLB.

Example:

```yaml
# HELP ohmyglb_gslb_healthy_records_total Number of healthy records observed by OhMyGLB.
# TYPE ohmyglb_gslb_healthy_records_total gauge
ohmyglb_gslb_healthy_records_total{name="test-gslb",namespace="test-gslb"} 6
```

#### `managed_hosts_total`

Number of managed hosts observed by OhMyGLB.

Example:

```yaml
# HELP ohmyglb_gslb_managed_hosts_total Number of managed hosts observed by OhMyGLB.
# TYPE ohmyglb_gslb_managed_hosts_total gauge
ohmyglb_gslb_managed_hosts_total{name="test-gslb",namespace="test-gslb",status="Healthy"} 1
ohmyglb_gslb_managed_hosts_total{name="test-gslb",namespace="test-gslb",status="NotFound"} 1
ohmyglb_gslb_managed_hosts_total{name="test-gslb",namespace="test-gslb",status="Unhealthy"} 2
```

Served on `0.0.0.0:8383/metrics` endpoint

### Custom resource specific metrics

Info metrics, automatically exposed by operator based on the number of the current instances of an operator's custom resources in the cluster.

Example:

```yaml
# HELP gslb_info Information about the Gslb custom resource.
# TYPE gslb_info gauge
gslb_info{namespace="test-gslb",gslb="test-gslb"} 1
```

Served on `0.0.0.0:8686/metrics` endoint

[prometheus]: https://prometheus.io/
[prometheus-operator]: https://github.com/coreos/prometheus-operator
[service-monitor]: https://github.com/coreos/prometheus-operator#customresourcedefinitions
[controller-runtime-metrics]: https://book.kubebuilder.io/reference/metrics.html
