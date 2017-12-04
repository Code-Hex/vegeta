INTERNAL_BIN_DIR=sub_bin
GOVERSION=$(shell go version)
GOOS=$(word 1,$(subst /, ,$(lastword $(GOVERSION))))
GOARCH=$(word 2,$(subst /, ,$(lastword $(GOVERSION))))
GO15VENDOREXPERIMENT=1
HAS_GLIDE:=$(shell which glide)


project = $(shell basename $(PWD))
server = ./app/$(project)
cli = ./app/cli
import = github.com/Code-Hex/$(project)
port = 8080
pid = $(PWD)/$(project).pid

proto-plugin:
	@protoc -I/usr/local/include -Iprotos \
			-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
			--go_out=plugins=grpc:protos collection.proto

# proto-gateway:
#	@protoc -I/usr/local/include -Iprotos \
#			-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#			--grpc-gateway_out=logtostderr=true:protos collection.proto

# proto-swagger:
#	@protoc -I/usr/local/include -Iprotos \
#			-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#			--swagger_out=logtostderr=true:protos collection.proto

proto: proto-plugin # proto-gateway proto-swagger

sass:
	@cd frontend && gulp sass

build:
	@go generate
	go build $(server)

build-cli:
	@go build $(cli)

run:
	@$(GOPATH)/bin/start_server --port=$(port) --pid-file=$(pid) -- ./$(project)

restart:
	@cat $(pid) | xargs kill -HUP

stop:
	@cat $(pid) | xargs kill -TERM

test: deps
	@PATH=$(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH):$(PATH) go test -v $(shell glide nv)

deps: glide
	@PATH=$(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH):$(PATH) glide install
	go get github.com/golang/lint/golint
	go get github.com/mattn/goveralls
	go get github.com/axw/gocov/gocov

$(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH)/glide:
ifndef HAS_GLIDE
	@mkdir -p $(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH)
	@curl -L https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-$(GOOS)-$(GOARCH).zip -o glide.zip
	@unzip glide.zip
	@mv ./$(GOOS)-$(GOARCH)/glide $(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH)/glide
	@rm -rf ./$(GOOS)-$(GOARCH)
	@rm ./glide.zip
endif

glide: $(INTERNAL_BIN_DIR)/$(GOOS)/$(GOARCH)/glide

lint: deps
	@for dir in $$(glide novendor); do \
	golint $$dir; \
	done;

cover: deps
	goveralls

.PHONY: test deps lint cover
