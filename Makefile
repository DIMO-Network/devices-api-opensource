.PHONY: all build buildstatic buildstatic-in-docker build-dockerfile deps deps-in-docker build-docker-image test help

IMAGE_NAME=CHANGE_ME # docker image name for output
SRC_PATH=CHANGE_ME # path of project relative to $GOPATH/src

THIS_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VENDOR_DIR=$(THIS_DIR)/vendor

IMAGE_VERSION = $(shell echo $${CI_BUILD_REF_NAME:=master}|sed 's/\//./') # get image tag from gitlab's CI_BUILD_REF_NAME env
UID = $(shell id -u)
GID = $(shell id -g)
USER = $(shell echo $${USER:=someuser})
SSH_PATH= $(shell if [ ! -z $$CI ]; then echo $$PWD/ssh; else echo $$HOME/.ssh; fi)
GO_FILES=$(shell find . -iname '*.go' -type f | grep -v /vendor/ |grep -v ".gen.go"| grep -v ".pb.go") # All the .go files, excluding vendor/
BUILDER_IMAGE_NAME=goprotobuild # Whatever docker image you want to use as builder
GO_PACKAGES=$(shell go list ./... | grep -v /vendor/)

define docker-run =
	make build-dockerfile
	docker run --rm -i \
		-v $$PWD:/go/src/$(SRC_PATH) \
		-v $(SSH_PATH):$$HOME/.ssh \
		-u $(UID):$(GID) \
		-v /etc/passwd:/etc/passwd:ro \
		-v /etc/group:/etc/group:ro $(BUILDER_IMAGE_NAME) make -C /go/src/$(SRC_PATH) $(1) $(2) $(3)
endef

default: help

build: ## Builds a dynamic linked binary
	go version
	go build

buildstatic: ## Builds a static binary
	@GO15VENDOREXPERIMENT=1 CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o main ./main.go
	
buildstatic-in-docker: ## Builds a static binary using a docker build environment
	$(call docker-run,"buildstatic")

build-dockerfile:				## Create dockerfile which acts as build environment
	@docker build --pull -f Dockerfile -t $(BUILDER_IMAGE_NAME) .

deps:					## Installs dependencies using glide
	@glide --home /tmp/ install

deps-in-docker:			## does deps in docker
	@$(call docker-run,"deps","proto")

build-docker-image: buildstatic-in-docker	## Builds the deployment docker image
	@docker build --pull -t $(IMAGE_NAME):$(IMAGE_VERSION) .

test:							## Run all tests
	@go test $(GO_PACKAGES)

vet:
	@go vet $(GO_PACKAGES)

fmt:
	@go fmt $(GO_PACKAGES)

push: ## Push the images to local and remote registry
	@docker push $(IMAGE_NAME):$(IMAGE_VERSION)


help:							## Show this help.
	@grep -e "^[a-zA-Z_-]*:" Makefile|awk -F'##' '{gsub(/[ \t]+$$/, "", $$1);printf "%-30s\t%s\n", $$1, $$2}'