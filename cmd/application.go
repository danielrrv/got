package cmd

import (
	"fmt"
	internal "github.com/danielrrv/got/internal"
	"os"
)

const (
	initName = "init"
	addName  = "add"
)

var (
	initArguments = []Arg{{
		Name:         "path",
		DefaultValue: "",
		Usage:        "Repository root tree path.",
	}}
)

func Execute() int {
	application := NewApplication()
	//commands.
	application.AddCommand(initName, initArguments, CommandInit)
	application.AddCommand(addName, nil, CommandAdd)

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
	fmt.Println(repo.Index.Entries)
	repo.Index.AddOrModifyEntries(repo, args)
	if err := repo.Index.Persist(repo); err != nil {
		panic(err)
	}
	fmt.Println(repo.Index.Entries[0])
	// // intern
	return 0

}
