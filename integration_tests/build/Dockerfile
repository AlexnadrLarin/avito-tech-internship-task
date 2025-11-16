FROM golang:1.24.5 AS build-stage

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd handlers && CGO_ENABLED=0 go test -c -o /build/handlers-test .

FROM golang:1.24.5-alpine AS test-runner

WORKDIR /tests

COPY --from=build-stage /build/handlers-test .

CMD ["./handlers-test", "-test.v"]

