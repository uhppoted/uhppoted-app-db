CMD   = ./bin/uhppoted-app-db
DIST  ?= development
DEBUG ?= --debug

.DEFAULT_GOAL := build-all
.PHONY: clean
.PHONY: update
.PHONY: update-release

SQLITE3  = ../runtime/uhppoted-app-db/sqlite3/acl.db
MSSQL    = sqlserver://sa:UBxNxrQiKWsjncow7mMx@localhost?database=uhppoted
MYSQL    = mysql://uhppoted:qwerty@/uhppoted
POSTGRES = postgresql://uhppoted:qwerty@localhost:5432/uhppoted

all: test      \
     benchmark \
     coverage

clean:
	go clean
	rm -rf bin

update:
	go get -u github.com/uhppoted/uhppote-core@main
	go get -u github.com/uhppoted/uhppoted-lib@main
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

build-all: build test vet lint
	mkdir -p dist/$(DIST)/linux
	mkdir -p dist/$(DIST)/arm
	mkdir -p dist/$(DIST)/arm7
	mkdir -p dist/$(DIST)/arm6
	mkdir -p dist/$(DIST)/darwin-x64
	mkdir -p dist/$(DIST)/darwin-arm64
	mkdir -p dist/$(DIST)/windows
	env GOOS=linux   GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/linux        ./...
	env GOOS=linux   GOARCH=arm64         GOWORK=off go build -trimpath -o dist/$(DIST)/arm          ./...
	env GOOS=linux   GOARCH=arm   GOARM=7 GOWORK=off go build -trimpath -o dist/$(DIST)/arm7         ./...
	env GOOS=linux   GOARCH=arm   GOARM=6 GOWORK=off go build -trimpath -o dist/$(DIST)/arm6         ./...
	env GOOS=darwin  GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/darwin-x64   ./...
	env GOOS=darwin  GOARCH=arm64         GOWORK=off go build -trimpath -o dist/$(DIST)/darwin-arm64 ./...
	env GOOS=windows GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/windows      ./...

release: update-release build-all
	find . -name ".DS_Store" -delete
	tar --directory=dist/$(DIST)/linux        --exclude=".DS_Store" -cvzf dist/$(DIST)-linux-x64.tar.gz    .
	tar --directory=dist/$(DIST)/arm          --exclude=".DS_Store" -cvzf dist/$(DIST)-arm-x64.tar.gz      .
	tar --directory=dist/$(DIST)/arm7         --exclude=".DS_Store" -cvzf dist/$(DIST)-arm7.tar.gz         .
	tar --directory=dist/$(DIST)/arm6         --exclude=".DS_Store" -cvzf dist/$(DIST)-arm6.tar.gz         .
	tar --directory=dist/$(DIST)/darwin-x64   --exclude=".DS_Store" -cvzf dist/$(DIST)-darwin-x64.tar.gz   .
	tar --directory=dist/$(DIST)/darwin-arm64 --exclude=".DS_Store" -cvzf dist/$(DIST)-darwin-arm64.tar.gz .
	cd dist/$(DIST)/windows && zip --recurse-paths ../../$(DIST)-windows-x64.zip . -x ".DS_Store"

publish: release
	echo "Releasing version $(VERSION)"
	gh release create "$(VERSION)" "./dist/$(DIST)-arm-x64.tar.gz"      \
	                               "./dist/$(DIST)-arm7.tar.gz"         \
	                               "./dist/$(DIST)-arm6.tar.gz"         \
	                               "./dist/$(DIST)-darwin-arm64.tar.gz" \
	                               "./dist/$(DIST)-darwin-x64.tar.gz"   \
	                               "./dist/$(DIST)-linux-x64.tar.gz"    \
	                               "./dist/$(DIST)-windows-x64.zip"     \
	                               --draft --prerelease --title "$(VERSION)-beta" --notes-file release-notes.md

debug: build
	$(CMD) --debug get-events --dsn "sqlite3://$(SQLITE3)"
	$(CMD) --debug get-events --dsn "$(MSSQL)"

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
	$(CMD) get-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACLx --table:log OperationsLog
	$(CMD) get-acl --dsn "sqlite3://$(SQLITE3)" --file "../runtime/uhppoted-app-db/get-acl.tsv"
	cat ../runtime/uhppoted-app-db/get-acl.tsv

sqlite3-get-acl-with-pin: build
	$(CMD) --debug get-acl --dsn "sqlite3://$(SQLITE3)" --with-pin

sqlite3-put-acl: build
	$(CMD) put-acl --file "../runtime/uhppoted-app-db/acl.tsv" --dsn "sqlite3://$(SQLITE3)" --table:ACL ACLx --table:log OperationsLog

sqlite3-put-acl-with-pin: build
	sqlite3 "$(SQLITE3)" 'delete from ACLx'
	$(CMD) put-acl --with-pin --file "../runtime/uhppoted-app-db/acl.tsv" --dsn "sqlite3://$(SQLITE3)"  --table:ACL ACLx --table:log OperationsLog
	sqlite3 "$(SQLITE3)" 'select * from ACLx'

sqlite3-load-acl: build
	$(CMD) load-acl --dsn "sqlite3://$(SQLITE3)"
	$(CMD) load-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL
	$(CMD) load-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL --table:audit Audit

sqlite3-load-acl-with-pin: build
	$(CMD) load-acl --with-pin --dsn "sqlite3://$(SQLITE3)"
	$(CMD) load-acl --with-pin --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL
	$(CMD) load-acl --with-pin --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL --table:audit Audit --table:log OperationsLog

sqlite3-store-acl: build
	sqlite3 "$(SQLITE3)" 'delete from ACLz'
	$(CMD) store-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACLz --table:log OperationsLog
	sqlite3 "$(SQLITE3)" 'select * from ACLz'

sqlite3-store-acl-with-pin: build
	sqlite3 "$(SQLITE3)" 'delete from ACLz'
	$(CMD) store-acl --with-pin  --dsn "sqlite3://$(SQLITE3)" --table:ACL ACLz
	sqlite3 "$(SQLITE3)" 'select * from ACLz'

sqlite3-compare-acl: build
	$(CMD) compare-acl --dsn "sqlite3://$(SQLITE3)"
	$(CMD) compare-acl --dsn "sqlite3://$(SQLITE3)" --table:ACL ACL --table:audit Audit --table:log OperationsLog

sqlite3-compare-acl-with-pin: build
	$(CMD) compare-acl --with-pin --dsn "sqlite3://$(SQLITE3)"

sqlite3-compare-acl-to-file: build
	$(CMD) compare-acl --with-pin --dsn "sqlite3://$(SQLITE3)" --file "../runtime/uhppoted-app-db/compare.rpt"
	cat ../runtime/uhppoted-app-db/compare.rpt

sqlite3-get-events: build
	$(CMD) get-events --dsn "sqlite3://$(SQLITE3)" --table:log OperationsLog

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
	$(CMD) compare-acl --dsn "$(MSSQL)" --table:ACL ACL --table:audit Audit

mssql-compare-acl-with-pin: build
	$(CMD) compare-acl --with-pin --dsn "$(MSSQL)"
	$(CMD) compare-acl --with-pin --dsn "$(MSSQL)" --table:ACL ACL --table:audit Audit

mssql-compare-acl-to-file: build
	$(CMD) compare-acl --with-pin --dsn "$(MSSQL)" --file "../runtime/uhppoted-app-db/compare.rpt"
	cat ../runtime/uhppoted-app-db/compare.rpt

mssql-load-acl: build
	$(CMD) load-acl --dsn "$(MSSQL)"
	$(CMD) load-acl --dsn "$(MSSQL)" --table:ACL ACL
	$(CMD) load-acl --dsn "$(MSSQL)" --table:ACL ACL --table:audit Audit

mssql-load-acl-with-pin: build
	$(CMD) load-acl --with-pin --dsn "$(MSSQL)"
	$(CMD) load-acl --with-pin --dsn "$(MSSQL)" --table:ACL ACL
	$(CMD) load-acl --with-pin --dsn "$(MSSQL)" --table:ACL ACL --table:audit Audit

mssql-store-acl: build
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "DELETE FROM ACLz"
	$(CMD) store-acl --dsn "$(MSSQL)" --table:ACL ACLz
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "SELECT * FROM ACLz"

mssql-store-acl-with-pin: build
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "DELETE FROM ACLz"
	$(CMD) store-acl --with-pin  --dsn "$(MSSQL)" --table:ACL ACLz
	mssql-cli -U sa -P UBxNxrQiKWsjncow7mMx -d uhppoted -Q "SELECT * FROM ACLz"

mssql-get-events: build
	$(CMD) get-events --dsn "$(MSSQL)" --table:log OperationsLog

mysql-get-acl: build
	$(CMD) get-acl --dsn "$(MYSQL)" --with-pin
#	$(CMD) get-acl --dsn "mysql://uhppoted:qwerty@tcp(127.0.0.1:3306)/uhppoted" 

mysql-put-acl: build
	$(CMD) put-acl --file "../runtime/uhppoted-app-db/acl.tsv" --dsn "$(MYSQL)" --table:ACL ACLx

mysql-compare-acl: build
	$(CMD) compare-acl --dsn "$(MYSQL)" --table:ACL ACL --table:audit Audit --with-pin

mysql-load-acl: build
	$(CMD) load-acl --dsn "$(MYSQL)" --table:ACL ACL --table:log OperationsLog

mysql-store-acl: build
	$(CMD) store-acl --dsn "$(MYSQL)" --table:ACL ACLz

mysql-get-events: build
	$(CMD) get-events --dsn "$(MYSQL)" --table:log OperationsLog --batch-size 5

postgres-get-acl: build
	$(CMD) --debug get-acl --dsn "$(POSTGRES)"
	$(CMD)         get-acl --dsn "$(POSTGRES)" --with-pin
	$(CMD)         get-acl --dsn "$(POSTGRES)" --table:ACL ACLx

postgres-put-acl: build
	$(CMD) put-acl --file "../runtime/uhppoted-app-db/acl.tsv" --dsn "$(POSTGRES)" --with-pin --table:ACL ACLx

postgres-compare-acl: build
	$(CMD) compare-acl --dsn "$(POSTGRES)" --table:ACL ACL --table:audit Audit --with-pin

postgres-load-acl: build
	$(CMD) load-acl --dsn "$(POSTGRES)" --table:ACL ACL --table:log OperationsLog

postgres-store-acl: build
	$(CMD) store-acl --dsn "$(POSTGRES)" --table:ACL ACLz

postgres-get-events: build
	$(CMD) get-events --dsn "$(POSTGRES)" --table:log OperationsLog --batch-size 5
