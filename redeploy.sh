#!/bin/bash
git reset --hard
git pull
chmod a+x ./redeploy.sh
docker build -t qrsync_server .
EXISTING_CONTAINER_ID=docker container ls --format "table {{.ID}}\t{{.Ports}}" -a | grep "4010->4010" | awk '{print $1}'
if [ -z EXISTING_CONTAINER_ID ]
then
    docker stop $EXISTING_CONTAINER_ID || true
fi

docker system prune -f
fuser -k 4010/tcp
docker run -d -p 4010:4010 --restart=always -it qrsync_server
