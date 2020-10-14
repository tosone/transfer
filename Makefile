app_name = transfer
Target   = ${app_name}

.PHONY: build
build:
	CGO_ENABLED=0 go build -v -o release/${Target}

.PHONY: release
release:
	CGO_ENABLED=0 go build -v -o release/${Target} -ldflags "-s -w"

.PHONY: lint
lint:
	golangci-lint run -v
