#!/bin/bash

containerName="nomikindiscord"
scriptroot="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
docker container rm $containerName -f
docker build -t $containerName $scriptroot
docker run -d --name $containerName $containerName
echo "Run \`docker logs --tail 50 $containerName\` to ensure setup was successful"

