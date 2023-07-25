FROM golang:1.20.6-alpine3.18 as build

ARG RELEASE

COPY . /app

WORKDIR /app

ENV CGO_ENABLED=0 RELEASE=$RELEASE

RUN ./make.sh build

FROM alpine:3.18

COPY --from=build /app/app /app

EXPOSE 8005
