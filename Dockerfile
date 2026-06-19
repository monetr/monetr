FROM --platform=$BUILDPLATFORM golang:1.26.4-trixie AS base_builder
WORKDIR /monetr
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      # renovate: datasource=deb depName=build-essential versioning=deb
      build-essential=12.12 \
      # renovate: datasource=deb depName=ca-certificates versioning=deb
      ca-certificates=20250419 \
      # renovate: datasource=deb depName=cmake versioning=deb
      cmake=3.31.6-2 \
      # gcc-x86-64-linux-gnu \ # Add these back to support arm64 hosts compiling amd64
      # libc6-dev-amd64-cross \
      # renovate: datasource=deb depName=gcc-aarch64-linux-gnu versioning=deb
      gcc-aarch64-linux-gnu=4:14.2.0-1 \
      # renovate: datasource=deb depName=libc6-dev-arm64-cross versioning=deb
      libc6-dev-arm64-cross=2.41-11cross1 \
      # renovate: datasource=deb depName=git versioning=deb
      git=1:2.47.3-0+deb13u1 \
      # renovate: datasource=deb depName=nodejs versioning=deb
      nodejs=20.19.2+dfsg-1+deb13u2 \
      # renovate: datasource=deb depName=npm versioning=deb
      npm=9.2.0~ds1-3 \
      # renovate: datasource=deb depName=wget versioning=deb
      wget=1.25.0-2 && \
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
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make release -B MONETR_BUILD_TYPE=container

FROM debian:13-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      # renovate: datasource=deb depName=tzdata versioning=deb
      tzdata=2026b-0+deb13u1 \
      # renovate: datasource=deb depName=ca-certificates versioning=deb
      ca-certificates=20250419 \
      # renovate: datasource=deb depName=locales-all versioning=deb
      locales-all=2.41-12+deb13u3 \
    && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN groupadd -g 1000 monetr && \
    useradd -rm -d /home/monetr -s /bin/bash -g monetr -u 1000 monetr
RUN mkdir -p /etc/monetr && chown -R monetr:monetr /etc/monetr
USER monetr
WORKDIR /home/monetr

EXPOSE 4000
VOLUME ["/etc/monetr"]
ENTRYPOINT ["/usr/bin/monetr"]
CMD ["serve"]

COPY --from=monetr_builder /monetr/build/monetr /usr/bin/monetr
