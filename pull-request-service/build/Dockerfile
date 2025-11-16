FROM golang:1.24.5 AS build-stage

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM build-stage AS test-stage
RUN go test -v ./...

FROM build-stage AS build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o pull-request-service-bin ./cmd/main/

FROM alpine:latest AS release-stage

WORKDIR /pull-request-service

COPY --from=build /build/pull-request-service-bin .

CMD ["./pull-request-service-bin"]
