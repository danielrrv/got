package main

import (
	"os"
	cmd "github.com/danielrrv/got/cmd"
)
//entry point
func main() {
	os.Exit(cmd.Execute())
}
