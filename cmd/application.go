package cmd

import (
	internal "github.com/danielrrv/got/internal"
	"os"
)

const (
	initName = "init"
	addName  = "add"
	statusName ="status"
)

var (
	initArguments = []Arg{{
		Name:         "path",
		DefaultValue: "",
		Usage:        "got add <path>...",
	}}
)

func Execute() int {
	application := NewApplication()
	//commands.
	application.AddCommand(initName, initArguments, CommandInit)
	application.AddCommand(addName, nil, CommandAdd)
	application.AddCommand(statusName, nil, CommandStatus)

	return application.Run()
}

func getOrDefault(v string, d string) string {
	if len(v) == 0 {
		return d
	} else {
		return v
	}

}

func CommandInit(app *Application, args []string) int {
	pwd, err := os.Getwd()
	if err != nil {
		app.Report(err)
	}
	path := getOrDefault(args[0], pwd)
	_, err = internal.FindOrCreateRepo(path)
	if err != nil {
		app.Report(err)
	}
	// config  :=make(map[string]interface{}, 0)
	// config["v"] = "sas"
	// repo.SetConfig(config)
	return 0
}

func CommandAdd(app *Application, args []string) int {
	repo, err := internal.FindOrCreateRepo(app.pwd)
	if err != nil {
		app.Report(err)
	}

	repo.Index.AddOrModifyEntries(repo, args)
	if err := repo.Index.Persist(repo); err != nil {
		panic(err)
	}
	return 0
}

func CommandStatus(app *Application, args []string) int{
	repo, err := internal.FindOrCreateRepo(app.pwd)
	if err != nil {
		app.Report(err)
	}
	repo.Status()
	return 0
}