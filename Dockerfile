FROM --platform=$BUILDPLATFORM golang:1.24.5 AS base_builder
WORKDIR /monetr
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      build-essential \
      ca-certificates \
      cmake \
      # gcc-x86-64-linux-gnu \ # Add these back to support arm64 hosts compiling amd64
      # libc6-dev-amd64-cross \
      gcc-aarch64-linux-gnu \
      libc6-dev-arm64-cross \
      git \
      nodejs=18.* \
      npm \
      wget && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN git config --global --add safe.directory /monetr

FROM base_builder AS monetr_builder
ARG REVISION
ARG RELEASE
ARG BUILD_HOST

# Multi platform
ARG TARGETOS
ARG TARGETARCH

ARG GOFLAGS
ENV GOFLAGS=$GOFLAGS
COPY . /monetr
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make release -B

FROM debian:12-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      tzdata \
      ca-certificates \
      locales-all \
    && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN useradd -rm -d /home/monetr -s /bin/bash -g root -G sudo -u 1000 monetr
RUN mkdir -p /etc/monetr && chown -R 1000:1000 /etc/monetr
USER monetr
WORKDIR /home/monetr

EXPOSE 4000
VOLUME ["/etc/monetr"]
ENTRYPOINT ["/usr/bin/monetr"]
CMD ["serve"]

# Build args need to be present in each "FROM"
ARG REVISION
ARG RELEASE

LABEL org.opencontainers.image.url=https://monetr.app
LABEL org.opencontainers.image.source=https://github.com/monetr/monetr
LABEL org.opencontainers.image.authors=elliot.courant@monetr.app
LABEL org.opencontainers.image.vendor="monetr"
LABEL org.opencontainers.image.licenses="FSL-1.1-MIT"
LABEL org.opencontainers.image.title="monetr"
LABEL org.opencontainers.image.description="monetr's budgeting application"
LABEL org.opencontainers.image.version=${RELEASE}
LABEL org.opencontainers.image.revision=${REVISION}

COPY --from=monetr_builder /monetr/build/monetr /usr/bin/monetr
