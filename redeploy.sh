#!/bin/bash
git reset --hard
git pull
chmod a+x ./redeploy.sh
docker build -t qrsync_server .
docker stop $(docker ps -q --filter ancestor=qrsync_server ) || true
docker system prune -f
fuser -k 4010/tcp
docker run -d -p 4010:4010 --restart=always -it qrsync_server
