package cmd

import (
	"fmt"
	"os"

	internal "github.com/danielrrv/got/internal"
)

const (
	initCommand = "init"
)

var (
	initArguments = []Arg{{
		Name:         "path",
		DefaultValue: "",
		Usage:        "Pass the path",
	}}
)

func getOrDefault(v string, d string) string {
	if len(v) == 0 {
		return d
	} else {
		return v
	}
}

func CommandInit(app *Application, args []string) int {
	pwd, err := os.Getwd() 
	if err!= nil {
		app.Report(err)
	}
	path := getOrDefault(args[0], pwd)
	repo, err := internal.FindOrCreateRepo(path)
	if err != nil {
		app.Report(err)
		app.Close(1)
	}
	fmt.Println(repo)
	return 0
}
