package main

import (
	mario "github.com/jeffbrennan/mario/cmd"
	"github.com/jeffbrennan/mario/cmd/mario"
	_ "github.com/jeffbrennan/mario/cmd/summarize"
)

func main() {
	mario.Execute()
}
