FROM archlinux:latest

ADD dist/ledger_linux_amd64/aurex /usr/local/bin/aurex

EXPOSE 3068

CMD ["aurex", "server", "start"]