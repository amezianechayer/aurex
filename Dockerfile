FROM alpine

RUN apk add --update-cache curl \
   && rm -rf /var/cache/apk/*
COPY corren /usr/local/bin/corren

EXPOSE 3068