FROM golang:1.20.3 as builder
RUN apt-get update
RUN apt-get install -y make sudo
RUN apt-get update && \
    apt-get install -y apt-transport-https ca-certificates gnupg curl

# Import the Google Cloud SDK package signing key
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
# Add the Google Cloud SDK repository to the apt sources list
RUN echo "deb https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

RUN apt-get update && \
    apt-get install -y google-cloud-sdk-gke-gcloud-auth-plugin

COPY . /src
WORKDIR /src
RUN rm -f go.sum
RUN go get ./...
RUN make linux-binary

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /src/bin/console-backend /app/console-backend

# copy the gke-gcloud-auth-plugin installed from builder image to path
COPY --from=builder /usr/bin/gke-gcloud-auth-plugin /usr/bin/gke-gcloud-auth-plugin

ENTRYPOINT ["/app/console-backend"]
