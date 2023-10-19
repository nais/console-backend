ARG GO_VERSION=1.21
FROM golang:${GO_VERSION}-alpine as builder
WORKDIR /src
COPY go.* /src/
RUN go mod download
COPY . /src
RUN go build -o bin/console-backend ./cmd/console-backend/main.go


FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /src/bin/console-backend /app/console-backend
ENTRYPOINT ["/app/console-backend"]
