FROM golang:stretch AS build-env

WORKDIR /go/src/github.com/epix/epix

RUN apt-get update -y
RUN apt-get install git -y

COPY . .

RUN make build

FROM golang:stretch

RUN apt-get update -y
RUN apt-get install ca-certificates jq -y

WORKDIR /root

COPY --from=build-env /go/src/github.com/epix/epix/build/epixd /usr/bin/epixd

EXPOSE 26656 26657 1317 9090

CMD ["epixd"]
