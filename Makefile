# Optional: prepend a docker CLI directory when it is not already on PATH.
#   make e2e DOCKER_BIN=/Applications/Docker.app/Contents/Resources/bin
ifdef DOCKER_BIN
export PATH := $(DOCKER_BIN):$(PATH)
endif

.PHONY: test e2e

test:
	go test ./pkg/containerkit/... ./pkg/client/...

e2e:
	go test -tags e2e -count=1 -timeout 15m -v ./e2e/...
