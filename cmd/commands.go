package cmd

import (
	"github.com/urfave/cli/v2"
)

type Command struct {
	AppCommands []*cli.Command
	Flags       *Flags
}

func (c *Command) SetCommands() {
	c.Flags = &Flags{}
	c.Flags.SetExisitingContractFlags()

	existingContractCommand := &cli.Command{
		Name:   "existing-contract",
		Flags:  c.Flags.ExistingContractFlags,
		Action: existingContractCmd,
	}

	newContractCommand := &cli.Command{
		Name:   "new-contract",
		Action: newContractCmd,
	}

	c.AppCommands = append(c.AppCommands, existingContractCommand, newContractCommand)
}
