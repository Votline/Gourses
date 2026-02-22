FROM golang:1.25.6-alpine AS builder
WORKDIR /build/users

COPY users/go.mod users/go.sum ./
RUN go mod download -x

COPY users ./
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o users

FROM alpine:3.18
WORKDIR /users

COPY --from=builder /build/users/users ./
CMD ["./users"]
