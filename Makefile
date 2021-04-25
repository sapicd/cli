.PHONY: help clean

BINARY=sapicli
VERSION=$(shell go run main.go -v)
CommitID=$(shell git log --pretty=format:"%h" -1)
Built=$(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
LINUX=$(BINARY).linux-amd64
MACOS=$(BINARY).darwin-amd64
WIN=$(BINARY).windows-amd64.exe
LDFLAGS=-ldflags "-s -w -X main.commitID=${CommitID} -X main.built=${Built}"

help:
	@echo "  make clean  - Remove binaries and vim swap files"
	@echo "  make gotool - Run go tool 'fmt' and 'vet'"
	@echo "  make build  - Compile go code and generate binary file"
	@echo "  make release- Format go code and compile to generate binary release"

gotool:
	go fmt ./
	go vet ./

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o $(LINUX) && chmod +x $(LINUX)

build-macos:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o $(MACOS) && chmod +x $(MACOS)

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o $(WIN) && chmod +x $(WIN)

build: clean gotool build-linux build-macos build-windows

release: gotool build
	mv $(LINUX) $(BINARY) && tar zcvf $(BINARY).$(VERSION)-linux-amd64.tar.gz $(BINARY) && rm $(BINARY)
	mv $(MACOS) $(BINARY) && tar zcvf $(BINARY).$(VERSION)-darwin-amd64.tar.gz $(BINARY) && rm $(BINARY)
	mv $(WIN) $(BINARY).exe && zip $(BINARY).$(VERSION)-windows-amd64.zip $(BINARY).exe && rm $(BINARY).exe

clean:
	find . -name '*.tar.gz' -exec rm -f {} +
	find . -name '*.zip' -exec rm -f {} +
	find . -name 'sapicli.*-amd64*' -exec rm -f {} +

docker:
	docker build -t staugur/sapicli .
