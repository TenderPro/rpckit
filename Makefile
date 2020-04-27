# application project makefile

SHELL          = /bin/bash

gen-ticker:
	docker run -ti --rm \
  -w $$PWD \
  -v $$PWD:$$PWD \
  tenderpro/protoc-go -I=./app/ticker \
    --gogofast_out=plugins=grpc,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types:app/ticker \
    ticker.proto

lint: ## Run linter
	@golangci-lint run ./...

## Run tests and fill coverage.out
cov: coverage.out

# internal target
coverage.out: $(SOURCES)
	$(GO) test -test.v -test.race -coverprofile=$@ -covermode=atomic -tags test ./...

## Open coverage report in browser
cov-html: cov
	$(GO) tool cover -html=coverage.out

## Clean coverage report
cov-clean:
	rm -f coverage.*
