FROM debian:12-slim AS base_builder
WORKDIR /build
RUN apt-get update && apt-get install -y git curl ca-certificates gnupg wget
RUN mkdir -p /etc/apt/keyrings && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg
RUN echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_18.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list
RUN apt-get update -y && \
    apt-get install -y \
    nodejs
RUN npm install -g pnpm
RUN wget -c https://golang.org/dl/go1.19.4.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz
ENV GOPATH /home/go
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH:


FROM base_builder AS source_code
RUN git config --global --add safe.directory /build
COPY . /build


FROM base_builder AS simple_icons
COPY --from=source_code /build /build
RUN git submodule update --init pkg/icons/sources/simple-icons


FROM base_builder AS node_modules
COPY --from=source_code ["/build/.npmrc", "/build/package.json", "/build/pnpm-lock.yaml", "/build/"]
RUN pnpm install


FROM base_builder AS go_modules
ARG GOFLAGS
ENV GOFLAGS=$GOFLAGS
COPY --from=source_code ["/build/go.mod", "/build/go.sum", "/build/"]
RUN go mod download


FROM go_modules AS third_party_notice
RUN apt-get install -y pkg-config cmake ruby-full libssl-dev
RUN gem install licensed -v 4.4.0
COPY --from=simple_icons /build /build
COPY --from=node_modules /build/node_modules /build/node_modules
RUN go mod download
RUN licensed cache
RUN licensed notices
RUN mkdir -p ./pkg/build && cat ./.licenses/monetr-API/NOTICE ./.licenses/monetr-UI/NOTICE > ./pkg/build/NOTICE.md


FROM node_modules AS ui_builder
COPY --from=source_code /build /build
ARG REVISION
ARG RELEASE
RUN RELEASE_VERSION=${RELEASE} RELEASE_REVISION=${REVISION} pnpm build --mode production


FROM go_modules AS monetr_builder
COPY --from=source_code /build /build
COPY --from=third_party_notice /build/pkg/build/NOTICE.md /build/pkg/build/NOTICE.md
COPY --from=ui_builder /build/pkg/ui/static /build/pkg/ui/static
COPY --from=simple_icons /build/pkg/icons/sources /build/pkg/icons/sources

# Build args need to be present in each "FROM"
ARG REVISION
ARG RELEASE
ARG BUILD_HOST

ARG GOFLAGS
ENV GOFLAGS=$GOFLAGS
RUN go build -ldflags "-s -w -X main.buildType=container -X main.buildHost=${BUILD_HOST:-`hostname`} -X main.buildTime=`date -u +"%Y-%m-%dT%H:%M:%SZ"` -X main.buildRevision=${REVISION} -X main.release=${RELEASE}" \
             -buildvcs=false \
             -o /usr/bin/monetr /build/pkg/cmd

FROM debian:12-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      tzdata=2023c-5 \
      ca-certificates=20230311 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

RUN useradd -rm -d /home/monetr -s /bin/bash -g root -G sudo -u 1000 monetr
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

COPY --from=monetr_builder /usr/bin/monetr /usr/bin/monetr
