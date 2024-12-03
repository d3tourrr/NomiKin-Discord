#!/bin/bash

# Change the value in quotation marks to change the default name for your companion
# You may want to do this if you're running multiple instances of this bot
# Ex: companionName = "friend_1"
containerName="nomikin-discord"
read -p "Docker container name is set to $containerName - is this okay? Press Enter to accept this name or enter another one: " inputName

if [ -n "$inputName" ]; then
  containerName=$inputName
fi

scriptroot="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
docker container rm $containerName -f
docker build -t $containerName $scriptroot
docker run -d --name $containerName
echo "Run `docker logs --tail 50 $containerName` to ensure setup was successful"
