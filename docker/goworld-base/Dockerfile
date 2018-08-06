FROM golang:1.10-alpine3.8

RUN apk add --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep
RUN go get -d github.com/xiaonanln/goworld
WORKDIR $GOPATH/src/github.com/xiaonanln/goworld
RUN dep ensure -v
RUN apk del git
CMD ["sh"]
