FROM golang:1.14.3-alpine AS build
WORKDIR /go/src/github.com/max-bytes/deploy-agent
COPY . /go/src/github.com/max-bytes/deploy-agent

# add ca-certificates, later copy them to final image
RUN apk update && apk --no-cache add ca-certificates && update-ca-certificates

 # git needed for fetching (internal) go modules from git repository
RUN apk add git

RUN go mod vendor
ENV CGO_ENABLED=0
ARG version
RUN go build -ldflags "-X main.version=$version" -o /out/deploy-agent ./cmd/deploy-agent

FROM scratch AS bin
COPY --from=build /out/deploy-agent /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./deploy-agent"]