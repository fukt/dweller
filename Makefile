.PHONY: all

all: init dep

init:
	go get github.com/golang/dep/cmd/dep

dep:
	dep ensure -v

codegen:
	vendor/k8s.io/code-generator/generate-groups.sh client,informer,lister,deepcopy github.com/fukt/dweller/pkg/client github.com/fukt/dweller/pkg/apis "dweller:v1alpha1"

build:
	cd cmd/dweller && go build -a -o ${CURDIR}/bin/dweller

docker:
	docker build -t fukt/dweller .

install:
	cd deploy && helm install --name dweller dweller

purge:
	helm delete --purge dweller

