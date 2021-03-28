FROM golang:1.16.2-alpine3.13 as builder
COPY ./ /build
WORKDIR /build
RUN go get ./...
RUN go build -o /bin/rest-api github.com/harderthanitneedstobe/rest-api/v0/cmd/api

FROM alpine:3.13.3

ARG REVISION

LABEL org.opencontainers.image.url=https://github.com/harderthanitneedstobe/rest-api
LABEL org.opencontainers.image.source=https://github.com/harderthanitneedstobe/rest-api
LABEL org.opencontainers.image.authors=me@elliotcourant.dev
LABEL org.opencontainers.image.revision=$REVISION
LABEL org.opencontainers.image.vendor="Harder Than It Needs To Be"
LABEL org.opencontainers.image.licenses="BSL-1.1"
LABEL org.opencontainers.image.title="REST API"
LABEL org.opencontainers.image.description="Harder Than It Needs To Be's REST API"

COPY --from=builder /bin/rest-api /usr/bin/rest-api
EXPOSE 4000
VOLUME ["/etc/harder"]
ENTRYPOINT ["/usr/bin/rest-api"]