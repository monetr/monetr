FROM nginx:1.19.6
EXPOSE 80
COPY ./build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
