# Go parameters
PROJECT_ROOT = $(shell pwd)

LD_FLAGS = '-linkmode "external" -extldflags "-Wl,-z,now"'

ROOT_DIR=$(shell git rev-parse --show-toplevel)
GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_DATE = $(shell date +"%F")
BUILD_TIME = $(shell date +"%T")
VERSION = V100R001C01

IMAGE_NAME = aass
ifneq ($(SERVICE_NAME),)
IMAGE_NAME = $(SERVICE_NAME)
endif

IMAGE_TAG = ${GIT_COMMIT}-${TIME}
ifneq ($(TAG),)
IMAGE_TAG = $(TAG)
endif

IMAGE_USER = mediacloud-mpc
ifneq ($(USER),)
IMAGE_UER = $(USER)
endif

IMAGE_REPOSITORY = registry-cbu.huawei.com
ifneq ($(REPOSITORY),)
IMAGE_REPOSITORY = $(REPOSITORY)
endif

IMAGE_FULL_NAME = ${IMAGE_REPOSITORY}/${IMAGE_USER}/${IMAGE_NAME}
$(info image full name is ${IMAGE_FULL_NAME}, tag is ${IMAGE_TAG})

ifdef BV
VERSION = $(BV)
endif
$(info BVersion is ${VERSION})

COMPILER = $(shell go version | awk '{ printf "%s-%s", $$3, $$4 }')

XLDFLAGS="-X ${GOMODULE}/boot.productVersion=${VERSION} \
    -X ${GOMODULE}/boot.compilerVersion=${COMPILER} \
    -X ${GOMODULE}/boot.gitBranch=${GIT_BRANCH} \
    -X ${GOMODULE}/boot.gitCommit=${GIT_COMMIT} \
    -X ${GOMODULE}/boot.buildDate=${BUILD_DATE} \
    -X ${GOMODULE}/boot.buildTime=${BUILD_TIME} \
    "

SERVICE = application-auto-scaling-service

ifeq ($(PAAS), 1.0)
VPATH = v1.0
else
VPATH = v2.0
endif

default: build

.PHONY: build test clean

build: clean
	CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2" go build -buildmode=pie -ldflags $(LD_FLAGS) -o bin/${SERVICE} $(PROJECT_ROOT)/cmd/application-auto-scaling-service/application_auto_scaling_service.go
	strip bin/${SERVICE}

clean:
	rm -rf build/*

docker: build
	@docker build -t ${IMAGE_NAME} ./

dockert: docker
	@docker tag ${IMAGE_NAME} ${IMAGE_FULL_NAME}:${IMAGE_TAG}
