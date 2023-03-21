CMD   = ./bin/uhppoted-app-db
DIST  ?= development
DEBUG ?= --debug

.PHONY: clean
.PHONY: update
.PHONY: update-release

all: test      \
     benchmark \
     coverage

clean:
	go clean
	rm -rf bin

update:
	go get -u github.com/uhppoted/uhppote-core@master
	go get -u github.com/uhppoted/uhppoted-lib@master
	go mod tidy

update-release:
	go get -u github.com/uhppoted/uhppote-core
	go get -u github.com/uhppoted/uhppoted-lib
	go mod tidy

format: 
	go fmt ./...

build: format
	mkdir -p bin
	go build -trimpath -o bin ./...

test: build
	go test ./...

vet: build
	go vet ./...


lint: build
	env GOOS=darwin  GOARCH=amd64 staticcheck ./...
	env GOOS=linux   GOARCH=amd64 staticcheck ./...
	env GOOS=windows GOARCH=amd64 staticcheck ./...

benchmark: build
	go test -bench ./...

coverage: build
	go test -cover ./...

build-all: test vet lint
	mkdir -p dist/$(DIST)/windows
	mkdir -p dist/$(DIST)/darwin
	mkdir -p dist/$(DIST)/linux
	mkdir -p dist/$(DIST)/arm
	mkdir -p dist/$(DIST)/arm7
	env GOOS=linux   GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/linux   ./...
	env GOOS=linux   GOARCH=arm64         GOWORK=off go build -trimpath -o dist/$(DIST)/arm     ./...
	env GOOS=linux   GOARCH=arm   GOARM=7 GOWORK=off go build -trimpath -o dist/$(DIST)/arm7    ./...
	env GOOS=darwin  GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/darwin  ./...
	env GOOS=windows GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/windows ./...

release: update-release build-all
	find . -name ".DS_Store" -delete
	tar --directory=dist --exclude=".DS_Store" -cvzf dist/$(DIST).tar.gz $(DIST)
	cd dist;  zip --recurse-paths $(DIST).zip $(DIST)

publish: release
	echo "Releasing version $(VERSION)"
	gh release create "$(VERSION)" \
	"./dist/uhppoted-app-db_$(VERSION).tar.gz" \
	"./dist/uhppoted-app-db_$(VERSION).zip" \
	--draft --prerelease --title "$(VERSION)-beta" --notes-file release-notes.md

debug: build
	$(CMD) version

godoc:
	godoc -http=:80 -index_interval=60s

usage: build
	$(CMD)

help: build
	$(CMD) help
	$(CMD) help commands
	$(CMD) help get-acl
	$(CMD) help load-acl
	$(CMD) help store-acl
	$(CMD) help compare-acl

version: build
	$(CMD) version

get-acl: build
	$(CMD) get-acl

get-acl-with-pin: build
	$(CMD) get-acl --with-pin

put-acl: build
	$(CMD) put-acl

load-acl-with-pin: build
	$(CMD) put-acl --with-pin

store-acl: build
	$(CMD) store-acl

store-acl-with-pin: build
	$(CMD) store-acl --with-pin

compare-acl: build
	$(CMD) compare-acl

compare-acl-file-with-pin: build
	$(CMD) compare-acl --with-pin


