FROM golang:1.10 as builder
RUN wget https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64
RUN mv dep-linux-amd64 /bin/dep && chmod +x /bin/dep
RUN apt-get update
RUN apt-get install -y protobuf-compiler
COPY . /go/src/github.com/WillAbides/xqsmee
RUN CGO_ENABLED=0 /go/src/github.com/WillAbides/xqsmee/script/build

FROM scratch as xqsmee
COPY --from=builder /go/src/github.com/WillAbides/xqsmee/bin/xqsmee .
ENTRYPOINT ["./xqsmee"]
