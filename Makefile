VER?=dev
COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS="-X internal/version.VERSION=$(VER) -X internal/version.COMMIT=$(COMMIT)"

app:
	go build -ldflags $(LDFLAGS) -o bin/lumina-server cmd/server/*.go
	go build -ldflags $(LDFLAGS) -o bin/lumina-agent cmd/agent/*.go
.PHONY: app

dashboard:
	cd dashboard && npm install && npm run build
.PHONY: dashboard

test:
	go test -v ./... -coverprofile cover.out
.PHONY: test

docs:
	swag init -g ./cmd/server/main.go ./cmd/agent/main.go -o docs
.PHONY: docs