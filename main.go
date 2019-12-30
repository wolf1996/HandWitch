package main

import (
	"github.com/wolf1996/HandWitch/cmd"
)

func main() {
	rootCmd := cmd.BuildAll()
	rootCmd.Execute()
}
