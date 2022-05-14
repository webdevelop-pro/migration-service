#!/bin/sh

PKG_LIST=$(go list ./... | grep -v /vendor/)
GO_FILES=$(find . -name '*.go' | grep -v _test.go)
WORK_DIR=$(pwd)
COMPANY=webdevelop-pro
SERVICE=migration-service

docker_build() {
  docker build -t ${COMPANY}/${SERVICE} .
}

docker_run() {
  docker stop ${SERVICE}
  docker rm ${SERVICE}
  docker run --name ${SERVICE} --env-file .example.env -d ${COMPANY}/${SERVICE}
}


build() {
  wget https://github.com/webdevelop-pro/migration-service/releases/download/0.1/app-0.1-linux-amd64.tar.gz
  tar xzf app-0.1-linux-amd64.tar.gz
  rm -rf app-0.1-linux-amd64.tar.gz
  chmod +x app
}

case $1 in

run)
  ./app
  ;;

build)
  build
  ;;

init)
  ./app --init
  ;;

help)
  cat make.sh | grep "^[a-z-]*)"
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

test-cloudbuild)
  export BRANCH_NAME=fake_branch
  export COMMIT_SHA=1234567
  cloud-build-local --dryrun=false .
  ;;
*)
  echo "unknown"
  ;;

esac