package main

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/WillAbides/xqsmee/cmd"
)

var (
	lineStartRe = regexp.MustCompile(`\n`)
	whiteLineRe = regexp.MustCompile(`\n\s+\n`)
)

const tpl = `/*
xqsmee (pronounced "excuse me") is a bit like https://github.com/probot/smee but with an eXtra Queue.

It's for some of the same situations where you would use smee, but you don't
want to miss events that are sent when you aren't watching.  A side effect
of using a queue is that even when multiple clients are watching, only one
client will get each request.

root usage:

%s

server usage:

%s

client usage:

%s

*/
package main
`

func main() {
	doc := fmt.Sprintf(tpl, cmdUsage(), cmdUsage("server"), cmdUsage("client"))
	doc = whiteLineRe.ReplaceAllString(doc, "\n\n")
	err := ioutil.WriteFile("doc.go", []byte(doc), 0644)
	if err != nil {
		panic(err)
	}
}

func indent(input string, indent string) string {
	input = indent + input
	return lineStartRe.ReplaceAllString(input, "\n"+indent)
}

func cmdUsage(commandLine ...string) string {
	bytes, err := cmd.CommandUsage(commandLine...)
	if err != nil {
		panic(err)
	}
	return indent(string(bytes), "  ")
}
