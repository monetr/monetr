FROM golang:1.17.4 as builder

ARG REVISION
ARG RELEASE
ARG GOFLAGS

WORKDIR /build

ENV GOFLAGS=$GOFLAGS
COPY ./go.mod ./go.sum /build/
RUN go mod download

COPY ./ /build
RUN go build -ldflags "-X main.buildRevision=$REVISION -X main.release=$RELEASE" -o /bin/monetr github.com/monetr/monetr/pkg/cmd

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y --no-install-recommends \
      tzdata=2021e-0ubuntu0.20.04  \
      ca-certificates=20210119~20.04.2 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*
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
LABEL org.opencontainers.image.version=$RELEASE

COPY --from=builder /bin/monetr /usr/bin/monetr
