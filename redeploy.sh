#!/bin/bash
git pull
chmod a+x ./redeploy.sh
docker build -t qrsync_server .
docker rm -f qrsync_server || true
docker system prune -f
fuser -k 4010/tcp
docker run -d -p 4001:4001 -it qrsync_server --restart=unless-stopped
