FROM node:16.2.0-buster AS builder

ARG ENVIRONMENT
ENV ENVIRONMENT=$ENVIRONMENT


COPY ./ /work
WORKDIR /work
RUN yarn install
RUN make build ENVIRONMENT=$ENVIRONMENT

FROM nginx:1.21.0
LABEL org.opencontainers.image.source=https://github.com/monetrapp/web-ui
EXPOSE 80
COPY --from=builder /work/build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
