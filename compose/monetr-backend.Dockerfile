# vim: set ft=dockerfile
FROM golang:1.24.8-trixie
RUN apt-get update && apt-get install -y locales-all tzdata
RUN groupadd -g ${GID:-1000} monetr
RUN useradd -rm -d /home/monetr -s /bin/bash -g root -g monetr -G sudo -u ${PID:-1000} monetr
USER monetr
WORKDIR /home/monetr
RUN mkdir -p /home/monetr/go/bin && mkdir /home/monetr/bin
ENV GOPATH=/home/monetr/go
ENV PATH="$PATH:/home/monetr/go/bin:/home/monetr/bin"
RUN go install github.com/cosmtrek/air@v1.29.0
RUN go install github.com/go-delve/delve/cmd/dlv@latest
