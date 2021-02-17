FROM golang:1.16.0 as builder
COPY ./ /build
WORKDIR /build
RUN go get ./...
RUN go build -o /bin/rest-api github.com/harderthanitneedstobe/rest-api/v0/cmd/api

FROM scratch
LABEL org.opencontainers.image.source=https://github.com/harderthanitneedstobe/rest-api
COPY --from=builder /bin/rest-api /usr/bin/rest-api
EXPOSE 4000
VOLUME ["/etc/harder"]
ENTRYPOINT ["/usr/bin/rest-api"]