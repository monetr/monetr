FROM golang:1.15.8 as builder
COPY ./ ./
RUN go get ./...
RUN go build -o /bin/rest-api github.com/harderthanitneedstobe/rest-api/v0/cmd/api

FROM scratch
COPY --from=builder /bin/rest-api /usr/bin/rest-api
RUN mkdir -p /etc/harder
EXPOSE 4000
VOLUME ["/etc/harder"]
ENTRYPOINT ["/usr/bin/rest-api"]