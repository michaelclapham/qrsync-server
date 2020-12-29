#!/bin/bash
git pull
go build -o qrsync_server
fuser -k 4010/tcp
./qrsync_server
#docker build -t qrsync_server .
#docker rm -f qrsync_server || true
#docker system prune -f
#docker run -d -p 4001:4001 -it qrsync_server --restart=unless-stopped
