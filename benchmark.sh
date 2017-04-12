#!/bin/bash

export GOPATH="`pwd`"

go test -v -bench=. -test.benchmem
