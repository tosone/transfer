app_name                = transfer
Target                  = ${app_name}

.PHONY: build
build:
	go build -v -o release/${Target}

.PHONY: release
release:
	go build -v -o release/${Target}${Suffix} -tags=jsoniter -ldflags "-s -w"

.PHONY: lint
lint:
	golangci-lint run -v
