FROM xiaonanln/goworld-base

RUN cd components/gate && go build
ENTRYPOINT ["components/gate/gate", "-gid"]
CMD ["1"]
