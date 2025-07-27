build:
	@go build -o bin/fs.exe

run: build
	@.\bin\fs.exe

test:
# 	@go test ./... -v
	@go test ./...