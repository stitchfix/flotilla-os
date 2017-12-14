FROM golang:latest

RUN mkdir -p /go/src/github.com/stitchfix/flotilla-os
ADD . /go/src/github.com/stitchfix/flotilla-os
RUN go get -u github.com/kardianos/govendor
WORKDIR /go/src/github.com/stitchfix/flotilla-os
RUN govendor sync
RUN go install github.com/stitchfix/flotilla-os

ENTRYPOINT /go/bin/flotilla-os /go/src/github.com/stitchfix/flotilla-os/conf
