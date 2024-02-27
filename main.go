package main

import (
	mario "github.com/jeffbrennan/mario/cmd"
	_ "github.com/jeffbrennan/mario/cmd/auth"
	_ "github.com/jeffbrennan/mario/cmd/summarize"
)

func main() {
	mario.Execute()
}
