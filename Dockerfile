ARG buildos=golang:1.18-alpine
ARG runos=scratch

# -- build dependencies with alpine --
FROM $buildos AS builder
WORKDIR /build
COPY . .
ARG goproxy
ARG TARGETARCH
RUN if [ "x$goproxy" != "x" ]; then go env -w GOPROXY=${goproxy},direct; fi ; \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags "-s -w -X main.built=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" .

# -- run application with a small image --
FROM $runos
COPY --from=builder /build/sapicli /bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["-i"]
ENTRYPOINT ["sapicli"]