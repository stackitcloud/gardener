#!/bin/bash
# v1.7.32: change in operation
# v1.7.33: chart api server vpn sidecar split podnetwork
# v1.7.34: controllermanager chart add node cidr ipv6
# v1.7.35: chart shoot-core vpn-shoot, podcidr split
# v1.7.36: add svc dual net
# v1.7.37: chart proxy configmap add podnetwork
# v1.7.38: chart vpn-shoot/apisrv, svccidr split
# v1.7.39: fix typo in service network shoot-core chart
# v1.7.40: changed validation of nodeip from ipv4 to ipv4,ipv6
# v1.7.41: set ipFamily of node exporter service to ipv4
# v1.7.42: more changes to dualstack also in mcm and os provider
# v1.7.43: changed dualStack feature gate field
# v1.7.44: added featureGate field to shoot v1beta1
# v1.7.45: changed from pointer to direct reference
# v1.7.46: added json definition
# v1.7.47: executed make generate
# v1.7.48: executed make generate v2
# v1.7.49: added split for node cidrs
# v1.7.50: changed ingress dns prefix to i
# v1.7.51: updated validation
# v1.7.52: lil fix
# v1.7.53: updated kube-controller-manager chart that it wont crash on empty featureGates
# v1.7.54: "
# v1.7.55: "
# v1.7.56: "
# v1.7.57: changed IngressPrefix
# v1.7.58: rebuild
# v1.7.59: introduced validation change for seed cidrs
# v1.7.60: introduced validation change for seed cidrs in gardenlet seed reconcile routine
# v1.7.61: fixed bug where b.Seed.Info.Spec.Networks.Nodes was not respected to be ipv4,ipv6
# v1.7.62: fixed bug where load balancer annotations were not accepted by shoot.gardener.cloud/use-as-seed annotation
# v1.10.2: rebased to v1.10.1
# v1.10.3: rebased to v1.7.62
# v1.10.4-ske: changed ingress spec for "networking.k8s.io/v1"
# v1.10.5-ske: changed service ipfamily for vpn-shoot to ipv4
# v1.10.6-ske: added cidr split for NODE_NETWORK in vpn-seed
# v1.10.7-ske: added cidr split for NODE_NETWORK in vpn-shoot
# v1.10.8-ske: changed ingress spec for "networking.k8s.io/v1" in seed-monitoring
# v1.10.9-ske: changed pathType Exact to Prefix in aggregate-prometheus ingress
# v1.10.10-ske: some rolebacks
# v1.10.11-ske: patched some ingress version stuff
# v1.10.12-ske: fixed silly stuff
# v1.10.13-ske: fixed silly stuff
# v1.10.14-ske: shoot.spec.networking.proxyConfig and containerd os-systemconfig
# v1.10.15-ske: changed kubelet path in hyperkube
# v1.10.16-ske: wrapped ExecStartPre of kubelet in sh -c
# v1.10.17-ske: Changed APIServer name to fqdn for MCM
# v1.10.18-ske: Changed APIServer DNS from fqdn to .svc, because APIServer Cert only valid for that
# v1.10.19-ske: added containerd runtime config
# v1.10.20-ske: fixed rendering bug
# v1.10.21-ske: added containerd RegistryEndpoint to worker config
# v1.10.22-ske: set InsecureSkipVerify protobuf to varint
# v1.10.23-ske: added proxy attributes for etcd backup  to controlplane.go, etcd.yaml, values.yaml
# v1.10.24-ske: Added null check for ProxyConfig in controlplane.go
# v1.10.25-ske: fixed silly bug
# v1.10.26-ske: added http_proxy to ctr image pull
# v1.10.27-ske: http_proxy to ctr image pull via shoot object
# v1.10.28-ske: http_proxy to ctr image pull via shoot object (with make generate)
# v1.10.29-ske: fixed silly bug
# v1.10.30-ske: added omitempty to shoot.spec.worker.cri.downloadHttpProxy
# v1.10.31-ske: added ipFamily to apiserver-service
# v1.10.32-ske: Add all IPv6 traffic to shoot api-server
# v1.10.33-ske: added componentResources to shoot.spec.provider
# v1.10.34-ske: add vpa config for gardenlet
# ... try to get it working
# v1.10.51-ske: add vpa config for gardenlet
# v1.10.53-ske: prometheus remote write
# v1.10.54-ske: fix remote write settings
# v1.17.0-ske: rebase on v1.17.0
# v1.17.0-ske-1: fixed node-cidr-mask-size-ipv4 for kube-controller-manager deployment in IPV6DualStack setup
# v1.17.0-ske-2: fixed gardener.cloud--allow-dns to support dual stack
# v1.17.0-ske-3: changed kubelet version check in checker.go
# v1.22.3-ske: rebase on v1.22.3
# v1.22.3-ske-1: vendor generate and import stuff

EFFECTIVE_VERSION=v1.22.3-ske-1
REGISTRY=registry.ske.eu01.stackit.cloud/gardener-ds
APISERVER_IMAGE_REPOSITORY=$REGISTRY/apiserver
CONROLLER_MANAGER_IMAGE_REPOSITORY=$REGISTRY/controller-manager
SCHEDULER_IMAGE_REPOSITORY=$REGISTRY/scheduler
SEED_ADMISSION_IMAGE_REPOSITORY=$REGISTRY/seed-admission-controller
GARDENLET_IMAGE_REPOSITORY=$REGISTRY/gardenlet

docker build --build-arg EFFECTIVE_VERSION=$EFFECTIVE_VERSION -t $APISERVER_IMAGE_REPOSITORY:$EFFECTIVE_VERSION         -f Dockerfile --target apiserver .
docker build --build-arg EFFECTIVE_VERSION=$EFFECTIVE_VERSION -t $CONROLLER_MANAGER_IMAGE_REPOSITORY:$EFFECTIVE_VERSION -f Dockerfile --target controller-manager .
docker build --build-arg EFFECTIVE_VERSION=$EFFECTIVE_VERSION -t $SCHEDULER_IMAGE_REPOSITORY:$EFFECTIVE_VERSION         -f Dockerfile --target scheduler .
docker build --build-arg EFFECTIVE_VERSION=$EFFECTIVE_VERSION -t $SEED_ADMISSION_IMAGE_REPOSITORY:$EFFECTIVE_VERSION    -f Dockerfile --target seed-admission-controller .
docker build --build-arg EFFECTIVE_VERSION=$EFFECTIVE_VERSION -t $GARDENLET_IMAGE_REPOSITORY:$EFFECTIVE_VERSION         -f Dockerfile --target gardenlet .

docker push $APISERVER_IMAGE_REPOSITORY:$EFFECTIVE_VERSION         
docker push $GARDENLET_IMAGE_REPOSITORY:$EFFECTIVE_VERSION
docker push $SEED_ADMISSION_IMAGE_REPOSITORY:$EFFECTIVE_VERSION
docker push $CONROLLER_MANAGER_IMAGE_REPOSITORY:$EFFECTIVE_VERSION
docker push $SCHEDULER_IMAGE_REPOSITORY:$EFFECTIVE_VERSION
