FROM golang:1.4.2
MAINTAINER maran@ethdev.com

RUN go get -u github.com/tools/godep
RUN go get -d github.com/ethereum/go-ethereum/cmd/geth
RUN go get -u gopkg.in/mgo.v2
RUN cd /go/src/github.com/ethereum/go-ethereum && godep restore
RUN apt-get update && apt-get install -y libgmp-dev
