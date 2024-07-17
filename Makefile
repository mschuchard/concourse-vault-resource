.PHONY: build

fmt:
	@go fmt ./...

tidy:
	@go mod tidy

get:
	@go get github.com/mschuchard/concourse-vault-resource

build: tidy
	@go build -o check cmd/check/main.go
	@go build -o in cmd/in/main.go
	@go build -o out cmd/out/main.go

release: tidy
	@go build -o check -ldflags="-s -w" cmd/check/main.go
	@go build -o in -ldflags="-s -w" cmd/in/main.go
	@go build -o out -ldflags="-s -w" cmd/out/main.go

bootstrap:
	@rm -f nohup.out
	# using cli for this avoids importing the entire vault/command package
	@nohup vault server -dev -dev-root-token-id="abcdefghijklmnopqrstuvwxyz09" &
	@go test -v -run TestBootstrap ./vault/util

shutdown:
	@killall vault

unit:
	@go test -v ./cmd ./concourse ./vault/...

accept:
	@go test -v ./cmd/check ./cmd/in ./cmd/out

resource:
	@docker build -t mschuchard/concourse-vault-resource -t mschuchard/concourse-vault-resource:${TAG} .
	@docker push -a mschuchard/concourse-vault-resource
