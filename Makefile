# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: geth android ios evm all test clean

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run
U = $(USER)
GO_DBG_BUILD = go build $(GO_FLAGS) -tags $(BUILD_TAGS),debug -gcflags=all="-N -l"  # see delve docs

geth:
	$(GORUN) build/ci.go install ./cmd/geth
	@echo "Done building."
	@echo "Run \"$(GOBIN)/geth\" to launch geth."

all:
	$(GORUN) build/ci.go install

test: all
	$(GORUN) build/ci.go test -coverage

lint: ## Run linters.
	$(GORUN) build/ci.go lint

evm-dbg:
	$(GO_DBG_BUILD) -o $(GOBIN)/ ./cmd/evm/...

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

run-geth-bsp:
	./build/bin/geth \
  --mainnet \
  --port 0 \
  --log.debug \
  --syncmode full \
  --datadir /Users/$(U)/Library/Ethereum/workshop/ \
  --replication.targets "redis://localhost:6379/?topic=replication" \
  --replica.result \
  --replica.specimen

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go install golang.org/x/tools/cmd/stringer@latest
	env GOBIN= go install github.com/fjl/gencodec@latest
	env GOBIN= go install github.com/golang/protobuf/protoc-gen-go@latest
	env GOBIN= go install ./cmd/abigen
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'
