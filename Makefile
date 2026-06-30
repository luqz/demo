BINARY  := add
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)
DIST    := dist

.PHONY: build test vet cross checksum release clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY)

test:
	go test -v ./...

vet:
	go vet ./...

cross:
	@mkdir -p $(DIST)
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-linux-amd64
	GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-linux-arm64
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-darwin-amd64
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -trimpath -o $(DIST)/add-windows-amd64.exe

checksum:
	cd $(DIST) && sha256sum add-* > SHA256SUMS

release: cross checksum

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)
