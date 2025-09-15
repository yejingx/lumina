VER?=dev
COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS="-X internal/version.VERSION=$(VER) -X internal/version.COMMIT=$(COMMIT)"

app:
	go build -ldflags $(LDFLAGS) -o bin/lumina main.go
.PHONY: app

run:
	go run main.go -c etc/config.yaml
.PHONY: run

dashboard:
	cd dashboard && npm install && npm run build
.PHONY: dashboard

test:
	go test -v ./... -coverprofile cover.out
.PHONY: test

docs:
	swag init -g ./main.go -o docs
.PHONY: docs