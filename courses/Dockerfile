FROM golang:1.25.6-alpine AS builder
WORKDIR /build/courses

COPY courses/go.mod courses/go.sum ./
RUN go mod download -x

COPY courses ./
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o courses

FROM alpine:3.18
WORKDIR /courses

COPY --from=builder /build/courses/courses ./
CMD ["./courses"]
