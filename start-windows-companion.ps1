$containerName = "NomiKinDiscord"
$inputName = Read-Host "Docker container name is set to $containerName - is this okay? Press Enter to accept this name or enter another one"

if (-not [string]::IsNullOrWhiteSpace($inputName)) {
    $containerName = $inputName
}

docker container rm $containerName -f
docker build -t $containerName $PSScriptRoot
docker run -d --name $containerName $containerName
Write-Output "Run `docker logs --tail 50 $containerName` to check and ensure setup was successful"
