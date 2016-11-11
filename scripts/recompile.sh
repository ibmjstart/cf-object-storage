GOOS=darwin GOARCH=amd64 go build -o ./binaries/darwin/cf-object-storage
GOOS=linux GOARCH=amd64 go build -o ./binaries/linux/cf-object-storage
GOOS=windows GOARCH=amd64 go build -o ./binaries/windows/cf-object-storage.exe
