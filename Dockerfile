FROM golang:1.10 as builder
COPY . /go/src/github.com/WillAbides/xqsmee
RUN CGO_ENABLED=0 GOOS=linux go install github.com/WillAbides/xqsmee/cmd/xqsmee

FROM scratch as xqsmee
ENV XQSMEE_MAXACTIVE=100
ENV XQSMEE_REDISURL="redis://redis:6379"
ENV XQSMEE_REDISPREFIX=xqsmee
ENV XQSMEE_HTTPADDR=":8000"
ENV XQSMEE_GRPCADDR=":9000"
COPY --from=builder /go/bin/xqsmee .
ENTRYPOINT ["./xqsmee"]
