FROM golang:alpine AS builder
WORKDIR /build/gateway

COPY gateway/go.mod gateway/go.sum ./
RUN go mod download -x

COPY gateway ./
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w"  -o gateway

FROM alpine:3.18
WORKDIR /gateway

COPY --from=builder /build/gateway/gateway ./
CMD ["./gateway"]
