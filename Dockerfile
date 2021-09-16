FROM golang:1.17.2 as builder

ARG REVISION
ARG BUILD_TIME
ARG RELEASE
ARG GOFLAGS

COPY ./ /build
WORKDIR /build

ENV GOFLAGS=$GOFLAGS
RUN go get ./...
RUN go build -ldflags "-X main.buildRevision=$REVISION -X main.buildtime=$BUILD_TIME -X main.release=$RELEASE" -o /bin/monetr github.com/monetr/monetr/pkg/cmd

FROM ubuntu:20.04

RUN apt-get update && apt-get install -y tzdata ca-certificates

EXPOSE 4000
VOLUME ["/etc/monetr"]
ENTRYPOINT ["/usr/bin/monetr"]
CMD ["serve"]

LABEL org.opencontainers.image.url=https://github.com/monetr/monetr
LABEL org.opencontainers.image.source=https://github.com/monetr/monetr
LABEL org.opencontainers.image.authors=elliot.courant@monetr.app
LABEL org.opencontainers.image.revision=$REVISION
LABEL org.opencontainers.image.vendor="monetr"
LABEL org.opencontainers.image.licenses="BSL-1.1"
LABEL org.opencontainers.image.title="monetr"
LABEL org.opencontainers.image.description="monetr's budgeting application"

COPY --from=builder /bin/monetr /usr/bin/monetr
