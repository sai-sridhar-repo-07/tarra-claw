package main

import (
	"fmt"
	"os"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
