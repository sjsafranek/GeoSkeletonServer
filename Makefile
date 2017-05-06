##=======================================================================##
## Makefile
## Created: Wed Aug 05 14:35:14 PDT 2015 @941 /Internet Time/
# :mode=makefile:tabSize=3:indentSize=3:
## Purpose:
##======================================================================##

SHELL=/bin/bash
PROJECT_NAME = GeoSkeletonServer
GPATH = $(shell pwd)

.PHONY: fmt get-deps update-deps test install build scrape clean

install: fmt get-deps
	@GOPATH=${GPATH} go build -o gskel_server server.go
	# @GOPATH=${GPATH} go build -o gskel_importer importer.go
	# @GOPATH=${GPATH} go build -o gskel_ts timeseries.go

build: fmt get-deps
	@GOPATH=${GPATH} go build -o gskel_server server.go
	# @GOPATH=${GPATH} go build -o gskel_importer importer.go
	# @GOPATH=${GPATH} go build -o gskel_ts timeseries.go

get-deps:
	mkdir -p "src"
	mkdir -p "pkg"
	mkdir -p "log"
	@GOPATH=${GPATH} go get github.com/boltdb/bolt
	@GOPATH=${GPATH} go get github.com/cihub/seelog
	@GOPATH=${GPATH} go get github.com/gorilla/mux
	@GOPATH=${GPATH} go get github.com/gorilla/websocket
	@GOPATH=${GPATH} go get github.com/paulmach/go.geojson
	@GOPATH=${GPATH} go get github.com/sjsafranek/DiffDB/diff_db
	@GOPATH=${GPATH} go get github.com/sjsafranek/DiffStore
	@GOPATH=${GPATH} go get github.com/sjsafranek/SkeletonDB
	@GOPATH=${GPATH} go get github.com/sjsafranek/GeoSkeletonDB

update-deps: get-deps
	@GOPATH=${GPATH} go get -u github.com/sjsafranek/DiffDB/diff_db
	@GOPATH=${GPATH} go get -u github.com/sjsafranek/DiffStore
	@GOPATH=${GPATH} go get -u github.com/sjsafranek/SkeletonDB
	@GOPATH=${GPATH} go get -u github.com/sjsafranek/GeoSkeletonDB

fmt:
	@GOPATH=${GPATH} gofmt -s -w ${PROJECT_NAME}
	@GOPATH=${GPATH} gofmt -s -w server.go
	# @GOPATH=${GPATH} gofmt -s -w importer.go
	# @GOPATH=${GPATH} gofmt -s -w timeseries.go

test:
	##./tcp_test.sh
	./benchmark.sh

scrape:
	@find src -type d -name '.hg' -or -type d -name '.git' | xargs rm -rf

clean:
	@GOPATH=${GPATH} go clean
