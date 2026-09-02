FROM --platform=$BUILDPLATFORM node:24.20.0-trixie-slim@sha256:50c3b2f6988dfc307b86e5301d69611af31f4789bdf232863b07d3b02fe55ae0 AS node

FROM --platform=$BUILDPLATFORM golang:1.26.6-trixie@sha256:b75d466dd608587fd66cca705a307ba65b889827d06ad61d6a75f0482b51b7c7 AS base_builder
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
      # Node links against libatomic on arm64.
      # renovate: datasource=deb depName=libatomic1 versioning=deb
      libatomic1=14.2.0-19 \
      # renovate: datasource=deb depName=wget versioning=deb
      wget=1.25.0-2 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Node comes out of the official image rather than apt so that we can pin an
# exact version, Debian trixie only ever ships Node 20 and cmake/FindPnpm.cmake
# already had to work around that. The upstream image verified the nodejs.org
# tarball against the Node release GPG keys and its SHA256 before unpacking it,
# so we inherit that. Same approach as compose/monetr-frontend.Dockerfile.
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

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

FROM debian:13-slim@sha256:3a39a0592364683e6bab97937b72cad5a8fa6dcbbee90edb3bb48c7f8e94f258
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
