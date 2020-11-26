## BUILD IMAGE
FROM  golang:1.15-alpine as builder

ARG git_tag
ARG git_commit

WORKDIR /go/src/github.com/nicholasanthonys/Golang-Body-Parser
COPY . .

RUN go mod vendor
RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/Golang-Body-Parser ./cmd/main.go

## FINAL IMAGE
FROM alpine:latest
WORKDIR /app/src
COPY --from=builder /go/src/github.com/nicholasanthonys/Golang-Body-Parser .

## Run build
CMD ["./build/Golang-Body-Parser"]
