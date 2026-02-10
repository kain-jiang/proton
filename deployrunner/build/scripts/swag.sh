#!/bin/bash

cd $(dirname $(readlink -f $0))/../..

go tool swag init -o api/rest/docs -g cmd/server/main.go  --parseDependency --parseInternal --exclude api/rest/docs
go tool swag fmt -g cmd/server/main.go --exclude api/rest/docs
