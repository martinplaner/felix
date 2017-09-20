FROM golang:alpine AS build-env

WORKDIR /go/src/github.com/martinplaner/felix

RUN apk update && apk add git

COPY . .
RUN go get -u -v github.com/ahmetb/govvv
RUN go test -v ./...
RUN GOOS=linux GOARCH=amd64 govvv build


FROM alpine

RUN apk update && apk upgrade

COPY --from=build-env /go/src/github.com/martinplaner/felix/felix /felix

RUN addgroup -g 6554 felix && adduser -s /bin/false -G felix -u 6554 -D felix
RUN mkdir -p /data
RUN chown felix:felix /data

VOLUME [ "/data" ]

USER felix
ENTRYPOINT [ "/felix", "-config", "/config.yml", "-datadir", "/data" ]
