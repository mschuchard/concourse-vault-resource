FROM golang:1.25-alpine AS build
WORKDIR /go/src/github.com/mschuchard/concourse-vault-resource
COPY . .
RUN apk add make && make release

FROM alpine:3.23
WORKDIR /opt/resource
COPY --from=build /go/src/github.com/mschuchard/concourse-vault-resource/check .
COPY --from=build /go/src/github.com/mschuchard/concourse-vault-resource/in .
COPY --from=build /go/src/github.com/mschuchard/concourse-vault-resource/out .
