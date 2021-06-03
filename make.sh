#!/bin/sh

PKG_LIST=`go list ./... | grep -v /vendor/`
GO_FILES=`find . -name '*.go' | grep -v _test.go`
WORK_DIR=`pwd`
COMPANY=webdevelop-pro
SERVICE=migration-service

docker_build() {
  docker build -t ${COMPANY}/${SERVICE} .
}

docker_run() {
  docker stop ${SERVICE}
  docker rm ${SERVICE}
  docker run --name ${SERVICE} --env-file .env.dev -d ${COMPANY}/${SERVICE}
}

build() {
  go build -ldflags "-s -w" -o app ./cmd/server/*.go &&
  chmod +x app
}

build_test() {
  go build -ldflags "-s -w" -o test-migration ./cmd/test-migration/*.go &&
  chmod +x test-migration
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
  CC=clang go test -msan -short ${PKG_LIST}
  ;;

coverage)
  mkdir /tmp/coverage > /dev/null
  rm /tmp/coverage/*.cov
  for package in ${PKG_LIST}; do
    go test -covermode=count -coverprofile "/tmp/coverage/${package##*/}.cov" "$package" ;
  done
  tail -q -n +2 /tmp/coverage/*.cov >> /tmp/coverage/coverage.cov
  go tool cover -func=/tmp/coverage/coverage.cov
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

build-test)
  build_test
  ;;

help)
  cat make.sh | grep "^[a-z-]*)"
  ;;

doc)
  for f in $PKG_LIST;
  do
    folder=${f/github.com\/${COMPANY}\/${SERVICE}\//./}
    godoc2md $folder > "${folder}/README.md"
  done
  ;;

docker-build)
  docker_build
  ;;

docker-run)
  docker_run
  ;;

gcloud-deploy)
  gcloud builds submit --tag gcr.io/${COMPANY}/${SERVICE}:dev
  ;;

*)
  echo "unknown"
  ;;

esac
