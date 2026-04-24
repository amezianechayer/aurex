#!/bin/sh
set -e

# Viper looks for config in $HOME/.corren and /etc/corren
# /etc/corren/corren.yaml is already in place from the Docker COPY
# CORREN_STORAGE_DIR and CORREN_SERVER_HTTP_BIND_ADDRESS env vars override defaults

exec corren server start
