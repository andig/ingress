.PHONY: default clean checks test build run mosquitto-clean publish-images test-release

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse --short HEAD)
VERSION := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))

BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')


default: clean checks test build

clean:
	rm -rf dist/ cover.out

checks:
	golangci-lint run

test: clean
	go test -v -cover ./...

build:
	@echo Version: $(VERSION) $(BUILD_DATE)
	go build -v -ldflags '-X "main.version=${VERSION}" -X "main.commit=${SHA}" -X "main.date=${BUILD_DATE}"' github.com/andig/ingress/cmd/ingress

run:
	go run -race github.com/andig/ingress/cmd/ingress

mosquitt-clean:
	./clean.sh

publish-images:
	@echo Version: $(VERSION) $(BUILD_DATE)
	seihon publish --version="$(TAG_NAME)" --image-name andig/ingress --base-runtime-image alpine --dry-run=false
	seihon publish --version="latest" --image-name andig/ingress --base-runtime-image alpine --dry-run=false

test-release:
	goreleaser --snapshot --skip-publish --rm-dist
