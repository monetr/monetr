# vim: ft=dockerfile
FROM ruby:3.2-bullseye
# Install cmake to build licensed, git for licensed to see dependencies.
RUN apt-get update && apt-get install -y cmake git curl
# TODO, Make renovate recognize this and auto update it.
RUN gem install licensed -v 4.4.0

# Install node and yarn for licensed.
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
RUN apt-get update -y && \
    apt-get install -y \
    nodejs
RUN npm install -g yarn

# Install golang for licensed.
RUN wget -c https://golang.org/dl/go1.19.1.linux-amd64.tar.gz
RUN tar -C /usr/local -xvzf go1.19.1.linux-amd64.tar.gz
ENV GOPATH /home/go
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH:

VOLUME /workspace
WORKDIR /workspace
ENTRYPOINT ["/bin/bash", "-c"]
