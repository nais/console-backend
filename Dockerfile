FROM golang:1.20.3 as builder
RUN apt-get update
RUN apt-get install -y make 
COPY go.* /src/
RUN go mod download
COPY . /src
WORKDIR /src
RUN make linux-binary

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /src/bin/console-backend /app/console-backend

ENTRYPOINT ["/app/console-backend"]
