package main

import (
	"os"

	"github.com/metalagman/semverctl/internal/cli"
)

func main() {
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
