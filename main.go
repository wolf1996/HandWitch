package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/cmd"
)

func main() {
	rootCmd, err := cmd.BuildAll()
	if err != nil {
		log.Fatalf("Failed to build comands %s", err.Error())
	}
	err = rootCmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute comand %s", err.Error())
	}
}
