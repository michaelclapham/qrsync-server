#!/bin/bash
git pull
docker build -t qrsync_server .
docker stop qrsync_server || true
docker system prune -f
docker run -d -p 4001:4001 -it qrsync_server --restart=unless-stopped
