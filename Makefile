-include .env

PUBLISH="$(shell pwd)/publish"
BUILD_DIR="./cmd/server"

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-s -w"

.PHONY: help linux darwin windows format clean prepare

all: help

## linux: Compile the binary with linux.
linux: prepare
	@cd ${BUILD_DIR}; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -trimpath -o ${PUBLISH}/${BINARY}-linux-${GOARCH} . ; \
	cd - >/dev/null

## darwin: Compile the binary with macos.
darwin: prepare
	@cd ${BUILD_DIR}; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -trimpath -o ${PUBLISH}/${BINARY}-darwin-${GOARCH} . ; \
	cd - >/dev/null

## windows: Compile the binary with windows.
windows: prepare
	@cd ${BUILD_DIR}; \
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -trimpath -o ${PUBLISH}/${BINARY}-windows-${GOARCH}.exe . ; \
	cd - >/dev/null

## format: format source codes
format:
	@cd ${BUILD_DIR}; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \
	cd - >/dev/null


prepare: clean
	@mkdir -p ${PUBLISH} ; \
	cp -rf config ${PUBLISH} ; \
	cp -rf ./shells/*.sh ${PUBLISH}

clean:
	@echo "  >  Cleaning build cache"
	@rm -rf ${PUBLISH}

help: Makefile
	@echo
	@echo " Choose a command run in "$(BUILD_DIR)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
