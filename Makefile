FILES?=./cmd
PLATFORM?=linux
ARCHITECTURE?=amd64

BINARY=event-timing-server
BUILDTIME=`date "+%F %T%Z"`
VERSION=`git describe --tags`

build:
	CGO_ENABLED=0 GOOS=$(PLATFORM) GOARCH=$(ARCHITECTURE) go build -ldflags="-X 'main.buildTime=$(BUILDTIME)' -X 'main.version=$(VERSION)' -s -w -extldflags '-static'" -o bin/$(BINARY) $(FILES)

install:
	go install github.com/go-bindata/go-bindata

generate: install
	go-bindata -pkg migrations -ignore bindata -nometadata -prefix athletes/migrations/ -o ./athletes/migrations/bindata.go ./athletes/migrations

run:
	go run $(FILES)

unit-test:
	go test `go list ./... | grep -v integration`

integration-test:
	go test `go list ./... | grep integration`

lint:
	golint -set_exit_status $(go list ./... | grep -v /vendor/)

cover:
	go test ./... -coverprofile cover.out
	go tool cover -html=cover.out

clean:
	rm -rf bin main
