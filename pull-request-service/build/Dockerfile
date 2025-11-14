FROM golang:1.24.5 AS build-stage

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o pull-request-service ./cmd/main/

FROM alpine:latest AS release-stage

WORKDIR /pull-request-service

COPY --from=build-stage /build/pull-request-service .

CMD ["./pull-request-service"]
