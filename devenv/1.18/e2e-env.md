export KO_DOCKER_REPO=kind.local
ko apply -f test/config
 ./test/upload-test-images.sh
k create ns serving-tests
go test -v -tags=hello -count=1 ./test/e2e
