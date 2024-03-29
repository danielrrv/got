package cmd


func Execute() int {
	application := NewApplication()
	//commands.
	application.AddCommand(initCommand, initArguments, CommandInit)

	return application.Run()
}
