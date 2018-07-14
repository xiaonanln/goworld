FROM xiaonanln/golang-jessie

RUN apt-get update && apt-get -y install sudo vim less

RUN groupadd -r goworld && useradd -m -g goworld -s /bin/bash goworld && echo goworld:goworld | chpasswd && adduser goworld sudo
# RUN chown goworld:goworld /home/goworld

USER goworld
ENV HOME=/home/goworld
ENV GOPATH=$HOME/go PATH=$HOME/go/bin:$PATH
RUN echo "GOPATH: $GOPATH"
RUN echo "PATH: $PATH"
WORKDIR $HOME

RUN go get -u github.com/golang/dep/cmd/dep
RUN go get -d github.com/xiaonanln/goworld
WORKDIR $GOPATH/src/github.com/xiaonanln/goworld
RUN dep ensure
RUN go get ./cmd/...
RUN make all

CMD ["/bin/bash"]
