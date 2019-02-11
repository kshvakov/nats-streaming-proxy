#!/bin/sh
set -e
if [ "$1" = remove ]; then
    /bin/systemctl stop    nats-streaming-proxy
    /bin/systemctl disable nats-streaming-proxy
    /bin/systemctl daemon-reload
fi