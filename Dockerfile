FROM golang:1.16.3-alpine3.13 as builder
COPY ./ /build
WORKDIR /build
RUN go get ./...
RUN go build -o /bin/rest-api github.com/monetrapp/rest-api/cmd/monetr

FROM alpine:3.13.5

RUN apk add --no-cache tzdata

ARG REVISION

LABEL org.opencontainers.image.url=https://github.com/monetrapp/rest-api
LABEL org.opencontainers.image.source=https://github.com/monetrapp/rest-api
LABEL org.opencontainers.image.authors=me@elliotcourant.dev
LABEL org.opencontainers.image.revision=$REVISION
LABEL org.opencontainers.image.vendor="monetr"
LABEL org.opencontainers.image.licenses="BSL-1.1"
LABEL org.opencontainers.image.title="REST API"
LABEL org.opencontainers.image.description="monetr's REST API"

COPY --from=builder /bin/rest-api /usr/bin/rest-api
EXPOSE 4000
VOLUME ["/etc/monetr"]
ENTRYPOINT ["/usr/bin/rest-api"]
