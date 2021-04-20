FROM golang:1.16.3 as builder
COPY ./ /build
WORKDIR /build
RUN go get ./...
RUN go build -o /bin/monetr github.com/monetrapp/rest-api/pkg/cmd

FROM debian:bullseye-slim

RUN apt-get install -y tzdata

ARG REVISION

LABEL org.opencontainers.image.url=https://github.com/monetrapp/rest-api
LABEL org.opencontainers.image.source=https://github.com/monetrapp/rest-api
LABEL org.opencontainers.image.authors=me@elliotcourant.dev
LABEL org.opencontainers.image.revision=$REVISION
LABEL org.opencontainers.image.vendor="monetr"
LABEL org.opencontainers.image.licenses="BSL-1.1"
LABEL org.opencontainers.image.title="REST API"
LABEL org.opencontainers.image.description="monetr's REST API"

COPY --from=builder /bin/monetr /usr/bin/monetr
EXPOSE 4000
VOLUME ["/etc/monetr"]
ENTRYPOINT ["/usr/bin/monetr", "serve", "--migrate=true"]
