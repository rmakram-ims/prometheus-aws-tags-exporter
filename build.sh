#!/bin/sh
set -ex

go mod tidy
go mod vendor
go build -o exporter