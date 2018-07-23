package main

import (
	"log"

	"github.com/WillAbides/xqsmee/cmd"
)

//go:generate go run ./script/docgen/docgen.go

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
