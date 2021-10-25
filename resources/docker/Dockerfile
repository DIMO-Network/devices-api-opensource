FROM golang:1.17 AS build

RUN useradd -u 10001 dimo

WORKDIR /go/src/github.com/DIMO-INC/devices-api/
COPY . /go/src/github.com/DIMO-INC/devices-api/

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOFLAGS=-mod=vendor

RUN make TAGS="timetzdata"

FROM busybox AS package

LABEL maintainer="DIMO <hello@dimo.zone>"

WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /go/src/github.com/DIMO-INC/devices-api/target/bin/devices-api .
COPY ./config/default.yaml /default.yaml

USER dimo

EXPOSE 8080

ENTRYPOINT ["/devices-api"]

CMD ["-c", "/default.yaml"]