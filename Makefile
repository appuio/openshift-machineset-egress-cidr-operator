OCP_VERSION := 4.7
K8S_VERSION := 1.20

test:
	go test -race ./...

lint:
	golangci-lint run -v ./...

deps:
	go get -d k8s.io/api@release-$(K8S_VERSION)
	go get -d k8s.io/apimachinery@release-$(K8S_VERSION)
	go get -d github.com/openshift/client-go@release-$(OCP_VERSION)
	go get -d github.com/openshift/api@release-$(OCP_VERSION)
	go get -d github.com/openshift/machine-api-operator@release-$(OCP_VERSION)
	go mod tidy
