.PHONY: build

fmt:
	@go fmt ./...

tidy:
	@go mod tidy

get:
	@go get github.com/mitodl/concourse-vault-resource

build: tidy
	@go build -o check cmd/check/main.go
	@go build -o in cmd/in/main.go
	@go build -o out cmd/out/main.go

release: tidy
	@go build -o check -ldflags="-s -w"  cmd/check/main.go
	@go build -o in -ldflags="-s -w"  cmd/in/main.go
	@go build -o out -ldflags="-s -w"  cmd/out/main.go

bootstrap:
	# using cli for this avoids importing the entire vault/command package
	@nohup vault server -dev -dev-root-token-id="abcdefghijklmnopqrstuvwxyz09" &
	@go test -v -run TestBootstrap ./vault

shutdown:
	@killall vault

unit:
	@go test -v ./...
