-include .env

PUBLISH="$(shell pwd)/publish"
BUILD_DIR="./msc_server/cmd"

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-s -w"

.PHONY: help linux darwin windows format clean mongo prepare

all: help

## linux: Compile the binary with linux.
linux: prepare
	@cd ${BUILD_DIR}; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -trimpath -o ${PUBLISH}/${BINARY}-linux-${GOARCH}-${VERSION} . ; \
	cd - >/dev/null

## darwin: Compile the binary with macos.
darwin: prepare
	@cd ${BUILD_DIR}; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -trimpath -o ${PUBLISH}/${BINARY}-darwin-${GOARCH}-${VERSION} . ; \
	cd - >/dev/null

## client: Compile client binary for your current platform
client: prepareC
	@cd ./msc_client/cmd ; \
	go build ${LDFLAGS} -trimpath -o ${PUBLISH}/cli/smscli . ; \
	cd - >/dev/null

## format: Format source codes
format:
	@cd ${BUILD_DIR}; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \
	cd - >/dev/null

## mongo: init mongodb
mongo:
	@cd ./auth_test ; \
	go test -v -run TestMongoStore_Load ./; \
	echo "Please set AuthClient.StoreType to \"mongo\" and set Mongo.URI in the config file." ; \
	cd - >/dev/null

prepare:
	@mkdir -p ${PUBLISH} ; \
	cp -rf msc_server/config ${PUBLISH} ; \
	cp -rf ./shells/*.sh ${PUBLISH}

prepareC:
	@mkdir -p ${PUBLISH}/cli ; \
	cp -rf msc_client/config ${PUBLISH}/cli ; \
	cp -rf ./shells/cstart.sh ${PUBLISH}/cli

clean:
	@echo "  >  Cleaning build cache"
	@rm -rf ${PUBLISH}

help: Makefile
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
