#!/bin/bash

export GOPATH="`pwd`"
cd GeoSkeletonServer
go test -v -bench=. -test.benchmem
