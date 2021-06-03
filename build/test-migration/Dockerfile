FROM golang:1.16.4-alpine3.13 as build

ARG RELEASE

COPY .. /app

WORKDIR /app

ENV CGO_ENABLED=0 RELEASE=$RELEASE

RUN ./make.sh build-test

FROM alpine:3.13

COPY --from=build /app/test-migration /test-migration

EXPOSE 8005

ENTRYPOINT [ "/test-migration" ]