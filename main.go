package main

import (
	log "github.com/sirupsen/logrus"

	"depositter/cmd"
)

func main() {
	// Basic app settings
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	app := cmd.NewApp()
	app.StartApp()

}
