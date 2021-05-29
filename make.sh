#!/bin/bash

PKG_LIST=`go list ./... | grep -v /vendor/`
GO_FILES=`find . -name '*.go' | grep -v _test.go`
WORK_DIR=`pwd`

function docker_build {
  docker build -t webdevelop-pro/migration-worker .
}

function docker_run {
  docker stop migration-worker
  docker rm migration-worker
  docker run --name migration-worker --env-file .env.dev -d webdevelop-pro/migration-worker
}

function docker_db {
  docker run --mount type=tmpfs,destination=${HOME}/tmpfs --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=test_webdevelop_pro -e POSTGRES_USER=postgres -p 5432:5432 -d postgres
}

function gcloud_db {
  docker run --mount type=tmpfs,destination=${HOME}/tmpfs --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=test_webdevelop_pro -e POSTGRES_USER=postgres -p 5432:5432 -d postgres
}

function docker_start {
  docker_build && 
  docker_run
}

function build {
  go build -ldflags "-s -w" -o app ./cmd/main/*.go &&
  chmod +x app
}

case $1 in

install)
  go install golang.org/x/lint/golint
  go install github.com/lanre-ade/godoc2md
  go install github.com/securego/gosec/cmd/gosec
  cp etc/pre-commit .git/hooks/pre-commit
  ;;

lint)
  golint -set_exit_status ${PKG_LIST}
  ;;

test)
  go test -count=1 ${PKG_LIST}
  ;;

race)
	go test -race -short ${PKG_LIST}
  ;;

memory)
	go test -msan -short ${PKG_LIST}
  ;;

coverage)
	./etc/coverage.sh;
  ;;

coverhtml)
	./etc/coverage.sh html;
  ;;

run)
	build
	./app
  ;;

gosec)
  echo "running gosec"
  gosec ./...
  ;;

build)
	build
  ;;

help)
	cat make.sh | grep "^[a-z]*)"
  ;;

doc)
  for f in $PKG_LIST;
  do
    folder=${f/github.com\/webdevelop-pro\/migration-worker\//./}
    godoc2md $folder > "${folder}/README.md"
  done
  ;;

docker-build)
  docker_build
  ;;

docker-start)
  docker_db;
  docker_build &&
  docker_run
  ;;

docker-db)
  docker_db
  ;;

docker-deploy)
  gcloud builds submit --tag gcr.io/webdevelop-pro/migration-worker:dev
  ;;

*)
  echo "unknown"
  ;;

esac
