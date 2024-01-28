package main

import (
	"flag"
	"jinx/pkg/util"
	"log"
)

func main() {
	flag.Parse()

	if len(flag.Args()) <= 0 {
		log.Fatal("no command line argument or options where provided")
	}

	command := flag.Arg(0)
	commandArgs := flag.Args()[1:]

	switch command {
	case util.START:
		HandleStart(commandArgs)
		break
	case util.STOP:
		HandleStop()
		break
	case util.RESTART:
		HandleRestart()
		break
	case util.DESTROY:
		HandleDestroy()
		break
	default:
		log.Fatalf("%s is an invalid or unrecognized command. valid commands are: start, stop, restart and destroy.", command)
	}
}
