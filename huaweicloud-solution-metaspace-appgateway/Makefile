export PROJECT_DIR=$(PWD)

LD_FLAGS = '-linkmode "external" -extldflags "-Wl,-z,now"'

default: build

.PHONY: build prepare test

prepare:
	@echo "exec prepare ..."
	@echo "GO111MODULE: $(GO111MODULE)"
	@echo "GONOSUMDB:   $(GONOSUMDB)"
	@echo "GOPROXY:	    $(GOPROXY)"
	@echo "GOSUMDB:	    $(GOSUMDB)"
	@echo -e "exec prepare done.\n"


build: prepare
	CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2" go build -buildmode=pie --ldflags $(LD_FLAGS) -o $(PROJECT_DIR)/bin/appgateway $(PROJECT_DIR)/cmd/application_gateway.go
	strip $(PROJECT_DIR)/bin/appgateway

clean:
	@rm -rf $(PROJECT_DIR)/bin/appgateway

