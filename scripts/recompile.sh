GOOS=darwin GOARCH=amd64 go build -o ./binaries/darwin/cf-object-storage
tar -cvzf ./binaries/zips/cf-object-storage-darwin.tar.gz ./binaries/darwin/cf-object-storage
GOOS=linux GOARCH=amd64 go build -o ./binaries/linux/cf-object-storage
tar -cvzf ./binaries/zips/cf-object-storage-linux.tar.gz ./binaries/linux/cf-object-storage
GOOS=windows GOARCH=amd64 go build -o ./binaries/windows/cf-object-storage.exe
tar -cvzf ./binaries/zips/cf-object-storage-windows.tar.gz ./binaries/windows/cf-object-storage.exe
