package main

import (
	"github.com/andrewwillette/andrewwillettedotcom/cmd"
	"github.com/andrewwillette/andrewwillettedotcom/log"
)

func main() {
	log.Configure()
	cmd.Execute()
}
