$containerName = "nomikindiscord"

docker container rm $containerName -f
docker build -t $containerName $PSScriptRoot
docker run -d --name $containerName $containerName
Write-Output "Run `docker logs --tail 50 $containerName` to check and ensure setup was successful"

