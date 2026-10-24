# Optional: prepend a docker CLI directory when it is not already on PATH.
#   make e2e DOCKER_BIN=/Applications/Docker.app/Contents/Resources/bin
ifdef DOCKER_BIN
export PATH := $(DOCKER_BIN):$(PATH)
endif

.PHONY: test e2e web containerkitd

test:
	go test ./pkg/kit/... ./pkg/client/... ./pkg/taskflow/...

web:
	cd web && npm install && npm run build

containerkitd: web
	go build -o bin/containerkitd ./cmd/containerkitd

e2e:
	go test -tags e2e -count=1 -timeout 15m -v ./e2e/...
