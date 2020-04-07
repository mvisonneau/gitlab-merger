package main

import (
	"github.com/mvisonneau/gitlab-merger/cli"
)

var version = ""

func main() {
	cli.Run(version)
}
