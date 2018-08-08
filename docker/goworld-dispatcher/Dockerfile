FROM xiaonanln/goworld-base

RUN cd components/dispatcher && go build
ENTRYPOINT ["components/dispatcher/dispatcher", "-dispid"]
CMD ["1"]
