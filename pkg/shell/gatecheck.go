package shell

type gatecheckCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the gatecheck CLI
// // shell: `gatecheck version`
func (s *gatecheckCmd) Version() *Command {
	return NewCommand(s.InitCmd().WithArgs("version"))
}
