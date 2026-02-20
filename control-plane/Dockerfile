FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY shared/go/ shared/go/
COPY control-plane/go.mod control-plane/go.sum ./control-plane/
RUN cd control-plane && go mod download

COPY control-plane/ control-plane/
RUN cd control-plane && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /qsgw-control-plane ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /qsgw-control-plane /usr/local/bin/qsgw-control-plane

EXPOSE 8085
ENTRYPOINT ["qsgw-control-plane"]
