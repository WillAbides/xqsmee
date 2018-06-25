xqsmee is like [smee](https://github.com/probot/smee) but with an e**X**tra **Q**ueue.

It's for some of the same situations where you would use smee, but you don't
want to miss events that are sent when you aren't watching.  A side effect
of using a queue is that even when multiple clients are watching, only one
client will get each request.

### example usage:

- prereq: you need to have a redis server running locally
- `go get -u github.com/WillAbides/xqsmee/cmd/*`
- each in a separate shell session:
    - start the server:
        `xqsmee`
    - start the client:
        `xqsmee-client -q "foo" -s "localhost:8000" --insecure`
    - post a message:
        `curl -d "hello world" http://localhost:8000/foo`

Each post you make to http://localhost:8000/foo will cause the client to
emit a json representation of the post.

```bash
$ xqsmee -h
Usage:
  xqsmee [flags]

Flags:
  -h, --help                              help for xqsmee
      --maxactive int                     max number of active redis connections (default 100)
      --redisprefix string                prefix for redis keys (default "xqsmee")
  -r, --redisurl string                   redis url (default "redis://:6379")
  -a, --tcp address to listen on string   tcp address to listen on (default ":8000")
```

```bash
$ xqsmee-client -h
Usage:
  xqsmee-client [flags]

Flags:
  -h, --help            help for xqsmee-client
      --ifs string      record separator (default "\n")
      --insecure        ignore ssl warnings
  -q, --queue string    xqsmee queue to watch
  -s, --server string   address of xqsmee server
```
