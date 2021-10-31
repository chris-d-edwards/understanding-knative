./kind-with-istio-knative.sh
kubectl apply -f https://github.com/n3wscott/sockeye/releases/download/v0.7.0/release.yaml
k apply -f  ./metrics/metrics-server-ignore-ssl.yaml



```
k apply -f - <<EOF
apiVersion: sources.knative.dev/v1beta2
kind: PingSource
metadata:
  name: test-ping-source
  namespace: default
spec:
  schedule: "*/1 * * * *"
  contentType: "application/json"
  data: '{"message": "Hello world!"}'
  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: sockeye
EOF

kubectl apply -f - <<EOF
apiVersion: sources.knative.dev/v1alpha1
kind: PrometheusSource
metadata:
  name: prometheus-source
spec:
  serverURL: http://prometheus.istio-system.svc.cluster.local:9090
  promQL: 'istio_requests_total'
  schedule: "* * * * *"
  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      namespace: default
      name: sockeye
EOF

http://prometheus.istio-system.svc.cluster.local:9090/api/v1/query?query=istio_requests_total
```

```
 cat <<EOF | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sleep
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sleep
  template:
    metadata:
      labels:
        app: sleep
    spec:
      containers:
      - name: sleep
        image: curlimages/curl
        command: ["/bin/sleep","3650d"]
        imagePullPolicy: IfNotPresent
EOF
```
