package cmd


func Execute() int {
	application := NewApplication()
	//commands.
	application.AddCommand(initCommandName, initArguments, CommandInit)

	return application.Run()
}
