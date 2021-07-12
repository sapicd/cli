# -- build dependencies with alpine --
FROM golang:1.16.3-alpine3.13 AS builder

LABEL maintainer=me@tcw.im

WORKDIR /build

COPY . .

ARG goproxy=https://goproxy.cn

RUN go env -w GOPROXY=${goproxy},direct

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.built=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" -o sapicli .

# run application with a small image
FROM scratch

COPY --from=builder /build/sapicli /bin/

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["-i"]

ENTRYPOINT ["sapicli"]
