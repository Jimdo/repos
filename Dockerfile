FROM golang:alpine

RUN apk --update add ca-certificates

ADD . /go/src/github.com/Jimdo/repos
WORKDIR /go/src/github.com/Jimdo/repos

RUN go install -v

ENTRYPOINT ["repos"]

EXPOSE 3000
