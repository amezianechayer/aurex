FROM alpine:latest

ADD aurex /usr/local/bin/aurex

EXPOSE 3068

CMD ["aurex", "server", "start"]