FROM golang:alpine3.8

LABEL VERSION="1.0"
LABEL DESCRIPTION="courier image."
LABEL MAINTAINER="mashenjun"
ENV GOPATH /go

RUN apk update && apk add curl bash tree tzdata \
    && cp -r -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && mkdir /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# copy binary
RUN mkdir -p $GOPATH/src/github.com/courier
COPY ./ $GOPATH/src/github.com/courier/
WORKDIR $GOPATH/src/github.com/courier
RUN go install main.go

ENV PATH=$PATH:$GOPATH/bin/
ENV TZ=Asia/Shanghai

CMD ["courier"]