CMD   = ./bin/uhppoted-app-db
DIST  ?= development
DEBUG ?= --debug

.DEFAULT_GOAL := build-all
.PHONY: clean
.PHONY: update
.PHONY: update-release

SQLITE3 = ../runtime/uhppoted-app-db/sqlite3/acl.db
MSSQL = sqlserver://sa:UBxNxrQiKWsjncow7mMx@localhost?database=uhppoted

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

benchmark: build
	go test -bench ./...

coverage: build
	go test -cover ./...

vet: build
	go vet ./...

lint: build
	env GOOS=darwin  GOARCH=amd64 staticcheck ./...
	env GOOS=linux   GOARCH=amd64 staticcheck ./...
	env GOOS=windows GOARCH=amd64 staticcheck ./...

vuln:
	govulncheck ./...

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
	$(CMD) compare-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL --table:audit Audit --with-pin
	$(CMD) load-acl    --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL --table:audit Audit

godoc:
	godoc -http=:80 -index_interval=60s

usage: build
	$(CMD)

help: build
	$(CMD) help
	$(CMD) help commands
	$(CMD) help get-acl
	$(CMD) help put-acl
	$(CMD) help load-acl
	$(CMD) help store-acl
	$(CMD) help compare-acl

version: build
	$(CMD) version

sqlite3-get-acl: build
	$(CMD) get-acl --dsn "sqlite3://$(SQLITE3)"
	$(CMD) get-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACLx
	$(CMD) get-acl --dsn "sqlite3://$(SQLITE3)" --file "../runtime/uhppoted-app-db/get-acl.tsv"
	cat ../runtime/uhppoted-app-db/get-acl.tsv

sqlite3-get-acl-with-pin: build
	$(CMD) --debug get-acl --dsn "sqlite3://$(SQLITE3)" --with-pin

sqlite3-put-acl: build
	sqlite3 "$(SQLITE3)" 'delete from ACLx'
	$(CMD) put-acl --file "../runtime/uhppoted-app-db/acl.tsv" --dsn "sqlite3://$(SQLITE3)" --table:ACL ACLx
	sqlite3 "$(SQLITE3)" 'select * from ACLx'

sqlite3-put-acl-with-pin: build
	sqlite3 "$(SQLITE3)" 'delete from ACLx'
	$(CMD) put-acl --with-pin --file "../runtime/uhppoted-app-db/acl.tsv" --dsn "sqlite3://$(SQLITE3)"  --table:ACL ACLx
	sqlite3 "$(SQLITE3)" 'select * from ACLx'

sqlite3-load-acl: build
	$(CMD) load-acl --dsn "sqlite3://$(SQLITE3)"
	$(CMD) load-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL
	$(CMD) load-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL --table:audit Audit

sqlite3-load-acl-with-pin: build
	$(CMD) load-acl --with-pin --dsn "sqlite3://$(SQLITE3)"
	$(CMD) load-acl --with-pin --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL
	$(CMD) load-acl --with-pin --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL --table:audit Audit

sqlite3-store-acl: build
	sqlite3 "$(SQLITE3)" 'delete from ACLz'
	$(CMD) store-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACLz
	sqlite3 "$(SQLITE3)" 'select * from ACLz'

sqlite3-store-acl-with-pin: build
	sqlite3 "$(SQLITE3)" 'delete from ACLz'
	$(CMD) store-acl --with-pin  --dsn "sqlite3://$(SQLITE3)" --table:ACL ACLz
	sqlite3 "$(SQLITE3)" 'select * from ACLz'

sqlite3-compare-acl: build
	$(CMD) compare-acl --dsn "sqlite3://$(SQLITE3)"
	$(CMD) compare-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL --table:audit Audit

sqlite3-compare-acl-with-pin: build
	$(CMD) compare-acl --with-pin --dsn "sqlite3://$(SQLITE3)"

sqlite3-compare-acl-to-file: build
	$(CMD) compare-acl --with-pin --dsn "sqlite3://$(SQLITE3)" --file "../runtime/uhppoted-app-db/compare.rpt"
	cat ../runtime/uhppoted-app-db/compare.rpt

mssql-get-acl: build
	$(CMD) --debug get-acl --dsn "$(MSSQL)"
	$(CMD)         get-acl --dsn "$(MSSQL)"
	$(CMD)         get-acl --dsn "$(MSSQL)" --table:ACL ACLx

mssql-get-acl-with-pin: build
	$(CMD) get-acl --dsn "$(MSSQL)" --with-pin

mssql-put-acl: build
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "DELETE FROM ACLx"
	$(CMD) put-acl --file "../runtime/uhppoted-app-db/acl.tsv" --dsn "$(MSSQL)" --table:ACL ACLx
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "SELECT * FROM ACLx"

mssql-put-acl-with-pin: build
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "DELETE FROM ACLx"
	$(CMD) put-acl --with-pin --file "../runtime/uhppoted-app-db/acl.tsv" --dsn "$(MSSQL)" --table:ACL ACLx
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "SELECT * FROM ACLx"

mssql-compare-acl: build
	$(CMD) compare-acl --dsn "$(MSSQL)"

mssql-compare-acl-with-pin: build
	$(CMD) compare-acl --with-pin --dsn "$(MSSQL)"

mssql-compare-acl-to-file: build
	$(CMD) compare-acl --with-pin --dsn "$(MSSQL)" --file "../runtime/uhppoted-app-db/compare.rpt"
	cat ../runtime/uhppoted-app-db/compare.rpt

mssql-load-acl: build
	$(CMD) load-acl --dsn "$(MSSQL)"
	$(CMD) load-acl --dsn "$(MSSQL)" --table:ACL ACL

mssql-load-acl-with-pin: build
	$(CMD) load-acl --with-pin --dsn "$(MSSQL)"
	$(CMD) load-acl --with-pin --dsn "$(MSSQL)" --table:ACL ACL

mssql-store-acl: build
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "DELETE FROM ACLz"
	$(CMD) store-acl --dsn "$(MSSQL)" --table:ACL ACLz
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "SELECT * FROM ACLz"

mssql-store-acl-with-pin: build
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "DELETE FROM ACLz"
	$(CMD) store-acl --with-pin  --dsn "$(MSSQL)" --table:ACL ACLz
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "SELECT * FROM ACLz"

