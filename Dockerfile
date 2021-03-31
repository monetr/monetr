FROM nginx:1.19.9
LABEL org.opencontainers.image.source=https://github.com/harderthanitneedstobe/web-ui
EXPOSE 80
COPY ./build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
