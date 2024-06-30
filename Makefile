GOPATH              := $(or $(GOPATH), $(HOME)/go)
GOLINT              := golangci-lint run
GOLINTCLEARCACHE	:= golangci-lint cache clean
MOCKERY             := $(GOPATH)/bin/mockery
GO_TEST_PARALLEL    := go test -parallel 4 -count=1 -timeout 30s
GOSTATIC            := go build -ldflags="-w -s"

$(MOCKERY):
	GOPATH=$(GOPATH) go install -mod=mod github.com/vektra/mockery/v2@latest
start:
	NV_REDIS_COMMON_ENABLE=true NV_IS_LOCAL=true NV_ENV=dev NV_SERVICE_PORT=80 go run main.go
clean:
	rm -rf ./out/main cpu.pprof mem.pprof
build: clean
	go mod tidy && go mod vendor && $(GOSTATIC) -o out/main ./cmd/cli/
lint:
	$(GOLINTCLEARCACHE) && $(GOLINT) -v ./...
test:
	$(GO_TEST_PARALLEL) ./... -v -coverprofile=cover.out && go tool cover -html=cover.out
mock: $(MOCKERY) #if: Interface will mock | dir: folder of interface | sn: name of mock struct
	$(MOCKERY) --name=$(if) --dir=$(dir) --structname=$(sn) --output=$(dir)/mocks
sqlboiler:
	sqlboiler -c ./configs/sqlboiler/conf.toml --wipe mysql
generate:
	go generate ./...