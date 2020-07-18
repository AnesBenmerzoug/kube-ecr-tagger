FROM golang:1.14-buster as build

RUN mkdir /tmp/src

WORKDIR /tmp/src

# We run this first in order to cache the modules and speedup the docker build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build-static

# Second stage
# We use busybox instead of scratch so that we can use the shell for debugging
FROM busybox:1.32.0

COPY --from=build /tmp/src/bin/kube-ecr-tagger /usr/local/bin/

RUN adduser kube-ecr -D

USER kube-ecr

CMD ["kube-ecr-tagger"]