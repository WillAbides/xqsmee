FROM golang:1.10 as builder
RUN wget https://github.com/gobuffalo/packr/releases/download/v1.11.1/packr_1.11.1_linux_amd64.tar.gz
RUN tar -xzf packr_1.11.1_linux_amd64.tar.gz && mv ./packr /bin/packr
COPY . /go/src/github.com/WillAbides/xqsmee
RUN CGO_ENABLED=0 /bin/packr install github.com/WillAbides/xqsmee

FROM scratch as xqsmee
COPY --from=builder /go/bin/xqsmee .
ENTRYPOINT ["./xqsmee"]
