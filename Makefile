$(shell mkdir -p bin)

windows:
	@GOOS=windows GOARCH=amd64 go build -o bin/dnsproxy.exe main.go

osx:
	@GOOS=darwin GOARCH=amd64 go build -o bin/dnsproxy main.go

linux:
	@GOOS=linux GOARCH=amd64 go build -o bin/dnsproxy main.go