ARG GO_VERSION=1.21
FROM golang:${GO_VERSION} as builder
WORKDIR /src
COPY go.* /src/
RUN go mod download
COPY . /src
RUN make linux-binary

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /src/bin/console-backend /app/console-backend
ENTRYPOINT ["/app/console-backend"]
