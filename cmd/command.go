package cmd

import (
	"flag"
	"fmt"
	"os"
)

// go build -o got  && sudo cp got /usr/bin
func usage() {
	format := `Usage: remote-terminal [OPTIONS] COMMAND [ARGS]...
	Author: Daniel Rodrigo Rodriguez Vergara <drodrigo678@gmail.com>
	Options:
		-v, --version            show the application's version and exit.
	`
	fmt.Fprintln(os.Stderr, format)
}
type Application struct {
	stdErr *os.File
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
	return &Application{
		commands: make([]Command, 0),
	}
}

func (a *Application) Run()  int{
	for _, cmd := range a.commands {
		fmt.Println(cmd.name)
		if os.Args[1] == cmd.name{
			return cmd.Run(a, os.Args[2:])
		}
	}
	fmt.Println("Unnoek comand")
	return 1
}

func(a * Application)Report(format error){
	fmt.Fprintln(a.stdErr, format)	
}

func (a * Application)Close(code int){
	os.Exit(code)
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
			for !cmd.Parsed() {
			}
			_args := make([]string, len(arguments))
			for i, arg := range arguments {
				_args[i] = *arg
			}
			return callback(app, _args)
		},
	})
}
