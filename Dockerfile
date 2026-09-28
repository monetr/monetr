FROM --platform=$BUILDPLATFORM node:24.21.0-trixie-slim@sha256:8ec5d7557396cfe32d21c3f9c13072355ceab22b584578ca4bb28af31120cffe AS node

FROM --platform=$BUILDPLATFORM golang:1.27.1-trixie@sha256:433790e515d27dc6003e847e644cc0af956985cf315c1c58a3b73ee2dd305183 AS base_builder
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

FROM debian:13-slim@sha256:a99cfc517144bc59b1978475ec53b46ecabec7e43635402ee5b77cc54cd1b20a
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      # renovate: datasource=deb depName=tzdata versioning=deb
      tzdata=2026c-0+deb13u1 \
      # renovate: datasource=deb depName=ca-certificates versioning=deb
      ca-certificates=20250419 \
      # renovate: datasource=deb depName=locales-all versioning=deb
      locales-all=2.41-12+deb13u4 \
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
