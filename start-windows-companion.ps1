# Change the value in quotation marks to change the default name for your companion
# You may want to do this if you're running multiple instances of this bot
# Ex: $companionName = $defaultName = "friend_1"
$containerName = "discord_companion"

### DO NOT EDIT BELOW THIS LINE ###
$inputName = Read-Host "Docker container name is set to $containerName - is this okay? Press Enter to accept this name or enter another one"

if (-not [string]::IsNullOrWhiteSpace($inputName)) {$containerName = $inputName}

docker container rm $containerName -f
docker build -t $containerName $psscriptroot
docker run -d --name $containerName
Write-Output "Run `docker logs --tail 50 $containerName to check to ensure setup was successful"
