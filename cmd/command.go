package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
)

// go build -o got  && sudo cp got /usr/bin
func usage() {
	format := `Usage:
    got <command> [<args>]
	command:
		init		Create an empty repository or find a existing one in path provided.
		add			Add files to stage area.
		status		Report the state of the working tree.
   `
	
	fmt.Fprintln(os.Stderr, format)
}
type Application struct {
	stdErr *os.File
	pwd string
	commands []Command
}

type Command struct {
	name string
	Run  func(a * Application, args []string) int
}

type Arg struct {
	Name         string
	DefaultValue string
	Usage        string
}



func NewApplication() *Application {
	pwd, err := os.Getwd() 
	if err !=  nil {panic(err)}
	return &Application{
		pwd: pwd,
		commands: make([]Command, 0),
	}
}

func (a *Application) Run()  int{
	for _, cmd := range a.commands {
		if len(os.Args) < 2{
			usage()
			os.Exit(1)
		}
		if os.Args[1] == cmd.name {
			return cmd.Run(a, os.Args[2:])
		}
	}
	usage()
	a.Report(errors.New("unknown command"))
	os.Exit(1)
	return 1
}

func(a * Application)Report(format error){
	fmt.Fprintln(a.stdErr, format)	
}


func (a *Application) AddCommand(name string, args []Arg, callback func(app * Application,args []string) int) {
	cmd := flag.NewFlagSet(name, flag.ContinueOnError)
	arguments := make([]*string, 0)
	for _, v := range args {
		ptrS := cmd.String(v.Name, v.DefaultValue, v.Usage)
		arguments = append(arguments, ptrS)
	}
	a.commands = append(a.commands, Command{
		name: name,
		Run: func(app * Application,args []string) int {
			cmd.Parse(args)
			_args := make([]string,0)
			for _, arg := range arguments {
				_args = append(_args, *arg) 
			}
		
			_args = append(_args, slices.Clone(cmd.Args())...)
			
			return callback(app, _args)
		},
	})
}
