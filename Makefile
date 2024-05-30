test:
	go test -v ./...

GOPROXY ?= direct

mod:
	GOPROXY=$(GOPROXY) GONOSUMDB=yunion.io/x go mod tidy
	GOPROXY=$(GOPROXY) GONOSUMDB=yunion.io/x go mod vendor -v
