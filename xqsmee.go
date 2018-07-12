package main

import (
	"log"

	"github.com/WillAbides/xqsmee/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
