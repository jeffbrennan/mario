package main

import (
	mario "github.com/jeffbrennan/mario/cmd"
	_ "github.com/jeffbrennan/mario/cmd/compare"
	_ "github.com/jeffbrennan/mario/cmd/config"
	_ "github.com/jeffbrennan/mario/cmd/summarize"
)

func main() {
	mario.Execute()
}
