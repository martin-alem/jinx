package main

import (
	"flag"
	"fmt"
	"jinx/pkg/util/constant"
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
	case constant.START:
		HandleStart(commandArgs)
		break
	case constant.STOP:
		HandleStop()
		break
	case constant.RESTART:
		HandleRestart()
		break
	case constant.DESTROY:
		HandleDestroy()
		break
	case constant.VERSION:
		fmt.Printf("Jinx Version %s", constant.VERSION_NUMBER)
		break
	default:
		log.Fatalf("%s is an invalid or unrecognized command. valid commands are: start, stop, restart and destroy.", command)
	}
}
