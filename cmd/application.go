package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	internal "github.com/danielrrv/got/internal"
)

const (
	initName   = "init"
	addName    = "add"
	statusName = "status"
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
	application.AddCommand("commit",nil, CommandCommit)
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
	//TODO: pass configuration of the repo initialization here.
	return 0
}

// CommandAdd is the handler for the "add" command.
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

// CommandStatus is the handler for the "status" command.
func CommandStatus(app *Application, args []string) int {
	repo, err := internal.FindOrCreateRepo(app.pwd)
	if err != nil {
		app.Report(err)
	}
	repo.Status()
	return 0
}

// CommandCommit is the handler for the "commit" command.
func CommandCommit(app *Application, args []string) int {
	//TODO: add message and author.
	repo, err := internal.FindOrCreateRepo(app.pwd)
	if err != nil {
		app.Report(err)
	}

	files := make([]string, 0)
	for _, entry := range repo.Index.Cache {
		files = append(files, entry.PathName)
	}
	//what it does: persist the repo index.
	repo.Index.Persist(repo)
	//what it does: create a map from the files.
	m := internal.CreateTreeFromFiles(repo, files)
	//what it does: create a tree from the map.
	tree := internal.FromMapToTree(repo, m, filepath.Base(repo.GotTree))
	//what it does: traverse the tree and write the objects to the disk.
	commit := internal.CreateCommit(repo, &tree, "some-message", "")
	//what it does: write the commit to the disk.
	hash, err := internal.WriteObject(repo, *commit, internal.CommitHeaderName)
	if err != nil {
		panic(err)
	}
	fmt.Println("Committed with hash:", hash)
	return 0
}
