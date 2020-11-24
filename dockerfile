
FROM golang:1.13-alpine as appbuild

ARG git_tag
ARG git_commit

ADD . /go/src/github.com/nicholasanthonys/Golang-Body-Parser
WORKDIR /go/src/github.com/nicholasanthonys/Golang-Body-Parser

RUN apk add build-base
RUN go mod vendor
RUN go build github.com/nicholasanthonys/Golang-Body-Parser
RUN go build -buildmode=plugin -o plugin/transform.so plugin/transform.go


FROM alpine
WORKDIR /app
COPY --from=appbuild /go/src/github.com/nicholasanthonys/Golang-Body-Parser .



CMD ["./Golang-Body-Parser"]
