package main

import (
	"fmt"
	"os"
)

const MIN_NUMBER_OF_IMPUT_PARAMETERS = 2

func main() {

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)

	cli := CLI()

	instance, err := cli.Build(os.Args[1:])

	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	err = instance.Run()

	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	// s := <-c
	// fmt.Println("Cleanup...", s)
	os.Exit(1)
}
