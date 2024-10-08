# Change the value in quotation marks to change the default name for your companion
# You may want to do this if you're running multiple instances of this bot
# Ex: $companionName = $defaultName = "friend_1"
$companionName = "discord_companion"

### DO NOT EDIT BELOW THIS LINE ###
$inputName = Read-Host "Companion Name (name of the Docker container) is set to $companionName - is this okay? Press Enter to accept this name or enter another one"

if (-not [string]::IsNullOrWhiteSpace($inputName)) {$companionName = $inputName}

docker container rm $companionName -f
docker build -t $companionName $psscriptroot
docker run -d --name $companionName -e COMPANION_NAME=$companionName $companionName
