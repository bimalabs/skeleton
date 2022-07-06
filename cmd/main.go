package main

import (
	"os"

	skeleton "app"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = append(args, "run")
	}

	command := args[0]
	option := ""
	if len(args) > 1 {
		option = args[1]
	}

	skeleton.Application(command).Run(option)
}
