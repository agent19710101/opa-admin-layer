BINARY := bin/opa-admin-layer

.PHONY: build test fmt run render validate preflight dispatch closeout

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/opa-admin-layer

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

run: build
	$(BINARY) serve -addr :8080

render: build
	$(BINARY) render -input deploy/examples/dev-spec.json

validate: build
	$(BINARY) validate -input deploy/examples/dev-spec.json

preflight:
	bash scripts/cycle/preflight.sh

dispatch:
	bash scripts/cycle/dispatch.sh

closeout:
	bash scripts/cycle/closeout.sh
