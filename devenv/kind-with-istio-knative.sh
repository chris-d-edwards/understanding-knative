#!/usr/bin/env bash

cat << EOF > clusterconfig-1.18.yaml 
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.18.8
  extraPortMappings:
  - containerPort: 31080
    hostPort: 80
  - containerPort: 31443
    hostPort: 443
EOF

kind create cluster --config clusterconfig-1.18.yaml

export KNATIVE_VERSION=v0.23.1
export CA_CERT_VERSION=v1.2.0

version=`istioctl version --remote=false`
if [ $version != "1.7.6" ]; then
  echo "wrong istio version"
  return 1
fi
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: istio-system
  labels:
    istio-injection: disabled
EOF
cat << EOF > ./istio-minimal-operator.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  values:
    global:
      proxy:
        autoInject: disabled
      useMCP: false
    gateways:
      istio-ingressgateway: 
        name: cluster-local-gateway
        runAsRoot: true
  addonComponents:
    pilot:
      enabled: true
    tracing:
      enabled: false
    kiali:
      enabled: false
    prometheus:
      enabled: false
  components:
    ingressGateways:
      - name: istio-ingressgateway
        enabled: true
      - name: cluster-local-gateway
        enabled: true
        label:
          istio: cluster-local-gateway
          app: cluster-local-gateway
        k8s:
          service:
            type: ClusterIP
            ports:
            - port: 15020
              name: status-port
            - port: 80
              name: http2
            - port: 443
              name: https
EOF
istioctl install -f istio-minimal-operator.yaml
cat << EOF > ./patch-ingressgateway-nodeport.yaml
spec:
  type: NodePort
  ports:
  - name: http2
    nodePort: 31080
    port: 80
    protocol: TCP
    targetPort: 80
EOF
kubectl patch service istio-ingressgateway -n istio-system --patch "$(cat ./patch-ingressgateway-nodeport.yaml)"

# Install Knative
kubectl apply --filename https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-crds.yaml
kubectl apply --filename https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-core.yaml
kubectl apply --filename https://github.com/knative/net-istio/releases/download/${KNATIVE_VERSION}/release.yaml

# Install knative eventing
#kubectl apply -f https://github.com/knative/eventing/releases/download/${KNATIVE_VERSION}/eventing-crds.yaml
#kubectl apply -f https://github.com/knative/eventing/releases/download/${KNATIVE_VERSION}/eventing-core.yaml
#kubectl apply -f https://github.com/knative/eventing/releases/download/${KNATIVE_VERSION}/in-memory-channel.yaml
#kubectl apply -f https://github.com/knative/eventing/releases/download/${KNATIVE_VERSION}/mt-channel-broker.yaml



# Install Cert Manager
kubectl create ns cert-manager
kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/${CA_CERT_VERSION}/cert-manager.yaml
kubectl wait --for=condition=available --timeout=600s deployment/cert-manager-webhook -n cert-manager

kubectl patch configmap/config-domain \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"127.0.0.1.nip.io":""}}'

