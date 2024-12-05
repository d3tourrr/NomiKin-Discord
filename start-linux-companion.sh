#!/bin/bash

containerName="NomiKinDiscord"
read -p "Docker container name is set to $containerName - is this okay? Press Enter to accept this name or enter another one: " inputName

if [ -n "$inputName" ]; then
  containerName=$inputName
fi

scriptroot="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
docker container rm $containerName -f
docker build -t $containerName $scriptroot
docker run -d --name $containerName $containerName
echo "Run \`docker logs --tail 50 $containerName\` to ensure setup was successful"
