FROM alpine:3.13
RUN mkdir -p /etc/harder
COPY ./bin/rest-api /usr/bin/rest-api
EXPOSE 4000
VOLUME ["/etc/harder"]
ENTRYPOINT ["/usr/bin/rest-api"]