FROM golang:1.17.2
# vi: ft=dockerfile

RUN apk update && apk add curl \
                          git \
                          protobuf \
                          bash \
                          make \
                          openssh-client && \
     rm -rf /var/cache/apk/*


RUN go get github.com/Masterminds/glide


CMD make