FROM debian:12-slim AS base_builder
WORKDIR /work
RUN apt-get update && apt-get install -y \
  git \
  curl \
  ca-certificates \
  gnupg \
  wget \
  pkg-config \
  cmake \
  ruby-full \
  libssl-dev \
  nodejs=18.* \
  npm

RUN npm install -g pnpm
RUN wget -c https://golang.org/dl/go1.20.1.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.20.1.linux-amd64.tar.gz
ENV GOPATH /home/go
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH:
RUN git config --global --add safe.directory /work

FROM base_builder AS monetr_builder
ARG REVISION
ARG RELEASE
ARG BUILD_HOST

ARG GOFLAGS
ENV GOFLAGS=$GOFLAGS
COPY . /work
RUN make monetr-release

FROM debian:12-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      tzdata=2023c-5+deb12u1 \
      ca-certificates=20230311 \
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
LABEL org.opencontainers.image.licenses="BSL-1.1"
LABEL org.opencontainers.image.title="monetr"
LABEL org.opencontainers.image.description="monetr's budgeting application"
LABEL org.opencontainers.image.version=${RELEASE}
LABEL org.opencontainers.image.revision=${REVISION}

COPY --from=monetr_builder /work/build/monetr /usr/bin/monetr
