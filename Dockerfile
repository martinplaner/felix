FROM golang AS build-env

WORKDIR /go/src/github.com/martinplaner/felix

COPY . .
RUN go get -u -v github.com/ahmetb/govvv
RUN go test -v ./...
RUN govvv build


FROM scratch
COPY --from=build-env /go/src/github.com/martinplaner/felix/felix /felix
ENTRYPOINT [ "/felix" ]
