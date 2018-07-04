FROM golang:1.10 as builder
COPY . /go/src/github.com/WillAbides/xqsmee
RUN CGO_ENABLED=0 go install github.com/WillAbides/xqsmee/cmd/xqsmee

FROM scratch as xqsmee
COPY --from=builder /go/bin/xqsmee .
ENTRYPOINT ["./xqsmee"]
