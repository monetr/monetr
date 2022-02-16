FROM docker.io/library/golang:1.17.6 as dependencies
WORKDIR /build

# Build args need to be present in each "FROM"
ARG GOFLAGS
ENV GOFLAGS=$GOFLAGS

COPY go.mod .
COPY go.sum .
RUN go mod download

FROM dependencies AS monetr_builder
COPY . /build

# Build args need to be present in each "FROM"
ARG REVISION
ARG RELEASE

ARG GOFLAGS
ENV GOFLAGS=$GOFLAGS
RUN go build -ldflags "-s -w -X main.buildTime=`date -u +"%Y-%m-%dT%H:%M:%SZ"` -X main.buildRevision=${REVISION} -X main.release=${RELEASE}" -o /usr/bin/monetr /build/pkg/cmd

FROM docker.io/library/debian:bookworm-20211201-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      tzdata=2021e-1  \
      ca-certificates=20211016 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*
EXPOSE 4000
VOLUME ["/etc/monetr"]
ENTRYPOINT ["/usr/bin/monetr"]
CMD ["serve"]

# Build args need to be present in each "FROM"
ARG REVISION
ARG RELEASE

LABEL org.opencontainers.image.url=https://github.com/monetr/monetr
LABEL org.opencontainers.image.source=https://github.com/monetr/monetr
LABEL org.opencontainers.image.authors=elliot.courant@monetr.app
LABEL org.opencontainers.image.vendor="monetr"
LABEL org.opencontainers.image.licenses="BSL-1.1"
LABEL org.opencontainers.image.title="monetr"
LABEL org.opencontainers.image.description="monetr's budgeting application"
LABEL org.opencontainers.image.version=${RELEASE}
LABEL org.opencontainers.image.revision=${REVISION}

COPY --from=monetr_builder /usr/bin/monetr /usr/bin/monetr
