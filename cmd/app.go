package cmd

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

type App struct {
	app *cli.App
}

func NewApp() *App {
	c := &Command{}
	a := &App{
		app: &cli.App{},
	}
	c.SetCommands()
	c.Flags.SetAppFlags()

	a.app.Flags = c.Flags.AppFlags
	a.app.Commands = c.AppCommands
	a.app.Name = "FTN Depositter"
	a.app.Usage = "Tool for FTN deposits"
	a.app.Version = "1.1"
	return a
}

func (a *App) StartApp() {
	err := a.app.RunContext(context.TODO(), os.Args)
	if err != nil {
		log.Fatalf("can't start app! err: %s", err)
	}
}
