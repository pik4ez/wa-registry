# Dmitry Belov' Wolt Assignment

Docker registry with Redis as a cache, persistent volume claim and a garbage collect cron job.

## TODO

* HTTPS.

## Approach

Infrastracture as code (IAC) is the basic idea. The `terratest` library is the core of the solution. It allows to run Kubernetes cluster, apply the manifests and perform the needed tests on top of that.

## Creating secrets

```
docker run --rm --entrypoint htpasswd registry:2.6.2 -Bbn USERNAME PASSWORD > htpasswd
kubectl create secret generic auth-secret --from-file=htpasswd
```

## Running tests

Setup k3d.

Run k3d cluster with the port 5000 exposed.

```
k3d cluster create wolt-assignment -p "5000:80@loadbalancer" --volume "/tmp/mock-repository-storage:/tmp/mock-repository-storage" --agents 1
```

Run tests.

```
cd ./test
go test -v -count=1 -timeout 30m -tags kubernetes
```

Run the specific test.

```
go test -v -tags kubernetes -run TestRegistry
```

## Manual testing of bin/ scripts

Setup k3d.

Run k3d cluster with the port 5000 exposed.

```
k3d cluster create wolt-assignment -p "5000:80@loadbalancer" --volume "/tmp/mock-repository-storage:/tmp/mock-repository-storage" --agents 1
```

Create test namespace.

```
kubectl create ns test-registry
```

Deploy manifests.

```
kubectl apply --namespace test-registry -f k8s
kubectl apply --namespace test-registry -f test/k8s
```

Run cleaner.

```
CONFIG=bin/clean/config.yaml REGISTRY_URL="http://localhost:5000" REGISTRY_USER="testuser" REGISTRY_PASSWORD="testpassword" ./bin/clean/main.py
```

Run garbage collector.

```
NAMESPACE=test-registry ./bin/garbage-collector.sh
```


## Further thoughts

* It's necessary to use a non-local storage of access mode ReadWriteMany in the production environment. The reason is multiple registry pods share this storage. Thus, test/k8s/persistent-volume.yaml is just a test implementation and mustn't be used in production.
* DNS may be used to set better name for the registry service.
* It's possible to improve the test of persistent volume mount by using the unique file path.
* There's no test for janitor cron job. It should be possible to create one.
* There's a reason to set up monitoring and alerting on top of /metrics endpoint. It's also possible to use [hooks](https://docs.docker.com/registry/configuration/#hooks) for email alerts.
* Templating and separation of environments are omitted in this implementation. Helm or Kustomize may be used for the task.
* There must be separate htpasswd file for production environment. Reliable storage such as Vault must be used for secrets.
* Stronger [http.secret](https://docs.docker.com/registry/configuration/#http) should be set for production environment.
* Tags deletion mechanism may be improved. Advantage of the existing solution: it uses API and thus, it should be pretty sustainable. Disadvantages: it may be slow; it always removes all the images with the same content, not only the specified tag.
* TLS is not implemented.
* Jenkinsfile hasn't been tested at all. It shouldn't work.
