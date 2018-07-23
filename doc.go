/*
xqsmee (pronounced "excuse me") is a bit like https://github.com/probot/smee but with an eXtra Queue.

It's for some of the same situations where you would use smee, but you don't
want to miss events that are sent when you aren't watching.  A side effect
of using a queue is that even when multiple clients are watching, only one
client will get each request.

root usage:

  Usage: xqsmee <command>

  Flags:
    --help    Show context-sensitive help.

  Commands:
    version
      show the xqsmee version

    server
      run a server

    client <server> <queue>
      run the client

  Run "xqsmee <command> --help" for more information on a command.

server usage:

  Usage: xqsmee server

  run a server

  Flags:
        --help                      Show context-sensitive help.

    -r, --redisurl=redis://:6379    redis url
        --maxactive=100             max number of active redis connections
        --no-tls                    don't use tls (serve unencrypted http and
                                    grpc)
        --httpaddr=":8443"          tcp address for http connections
        --grpcaddr=":9443"          tcp address for grpc connections
        --redisprefix="xqsmee"      prefix for redis keys
        --tlskey=STRING             file containing a tls key
        --tlscert=STRING            file containing a tls certificate
        --publicurl="https://localhost:8443"
                                    the http url that end users will use

client usage:

  Usage: xqsmee client <server> <queue>

  run the client

  Arguments:
    <server>    server ip or dns address
    <queue>     xqsmee queue to watch

  Flags:
        --help         Show context-sensitive help.

    -p, --port=9443    server grpc port
        --insecure     don't check for valid certificate
        --no-tls       don't use tls (insecure)
        --ifs="\n"     record separator

*/
package main
