# Understanding-playground-knative

Simple repo to understand knative serving. Created this repo to investigate on a possible bug on knative readiness  handle exec by probes. with http probes it works for me  [issue](https://github.com/knative/serving/issues/11693)


# Prepare env 
Prequisite to prepare env with using the below script is necessary have installed  [Kind](https://kind.sigs.k8s.io/) , helm 3 or greater and have istiocl 1.7.6 in your PATH 

```sh
cd devenv
./kind-with-istio-knative.sh
```

# Prepare image

Created a custom images for add jq and a custom jar for descriminate if possible readiness is only for ``handle exec probes`` istead swap using a http probe problem doesn't exists


```sh
cd image
docker build -t kind.local/trino-356:356-jq .
kind load docker-image kind.local/trino-356:356-jq
```

# Testing

```sh
wget https://repo1.maven.org/maven2/io/trino/trino-cli/356/trino-cli-356-executable.jar -O trino-cli.jar
chmod +x trino-cli.jar
```

```sh
cd ./trino/trino-kserving-helm
helm install trinodb .
```

Wait Trinodb scaledown from the folder where trino-cli is installed

```sql
./trino-cli.jar --server http://trinodb-coordinator.default.127.0.0.1.nip.io
select activecount from jmx.current."trino.failuredetector:name=HeartbeatFailureDetector";
```
