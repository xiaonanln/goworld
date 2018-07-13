FROM xiaonanln/golang-jessie
RUN go get -d github.com/xiaonanln/goworld
RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR $GOPATH/src/github.com/xiaonanln/goworld
RUN dep ensure
RUN go get ./cmd/...
RUN make all
