package cmd

import (
	"os"

	internal "github.com/danielrrv/got/internal"
)

const (
	initCommandName = "init"
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
		
	}
	var config interface{}
	repo.SetConfig(config)
	return 0
}
