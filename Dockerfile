FROM         golang:1.4-alpine
MAINTAINER   Alvaro Saurin <alvaro.saurin@gmail.com>

RUN          apk add --update build-base git
RUN          mkdir -p /go/src/app
WORKDIR      /go/src/app
COPY         .         /go/src/app
RUN          make deps
RUN          make image                         && \
             cp prog/docker-env /usr/local/bin  && \
             rm -rf /go                            \
                    /usr/local/go                  \
                    /var/cache/apk/*

CMD ["/usr/local/bin/docker-env"]
