GOOS=darwin GOARCH=amd64 go build
mv cf-large-objects ./binaries/darwin
GOOS=linux GOARCH=amd64 go build
mv cf-large-objects ./binaries/linux
GOOS=windows GOARCH=amd64 go build
mv cf-large-objects.exe ./binaries/windows
