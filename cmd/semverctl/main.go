package main

import (
	"os"

	"github.com/metalagman/semverctl/internal/cli"
)

var execute = cli.Execute
var exit = os.Exit

func main() {
	err := execute()
	if err != nil {
		exit(1)
	}
}
