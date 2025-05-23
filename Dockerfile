FROM golang:1.23-alpine as build
WORKDIR /go/src/github.com/mschuchard/concourse-vault-plugin
COPY . .
RUN apk add make && make release

FROM alpine:3.21
WORKDIR /opt/resource
COPY --from=build /go/src/github.com/mschuchard/concourse-vault-plugin/check .
COPY --from=build /go/src/github.com/mschuchard/concourse-vault-plugin/in .
COPY --from=build /go/src/github.com/mschuchard/concourse-vault-plugin/out .
