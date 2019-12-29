FROM golang:1.13-buster as build

COPY . /tmp/src

WORKDIR /tmp/src

RUN make build-static

FROM busybox:1.31.1-glibc

COPY --from=build /tmp/src/bin/* /usr/local/bin/

RUN adduser kube-ecr -D

USER kube-ecr

CMD ["kube-ecr-tagger"]