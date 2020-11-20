
FROM golang:1.13-alpine

ARG git_tag
ARG git_commit

ADD . /go/src/github.com/nicholasanthonys/Golang-Body-Parser
WORKDIR /go/src/github.com/nicholasanthonys/Golang-Body-Parser

RUN apk add build-base
RUN go mod vendor
RUN go build github.com/nicholasanthonys/Golang-Body-Parser


FROM alpine
WORKDIR /usr/bin/Golang-Body-Parser
COPY --from=0 /go/src/github.com/nicholasanthonys/Golang-Body-Parser .


CMD ["./Golang-Body-Parser"]