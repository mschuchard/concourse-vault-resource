FROM golang:1.21-alpine as build
WORKDIR /go/src/github.com/mitodl/concourse-vault-plugin
COPY . .
RUN apk add make && make release

FROM alpine:3.19
WORKDIR /opt/resource
COPY --from=build /go/src/github.com/mitodl/concourse-vault-plugin/check .
COPY --from=build /go/src/github.com/mitodl/concourse-vault-plugin/in .
COPY --from=build /go/src/github.com/mitodl/concourse-vault-plugin/out .
