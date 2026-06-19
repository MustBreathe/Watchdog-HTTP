package app

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const SERVER = "server"
const MONITOR = "monitor"

type CommandLineInterface struct {
	strategy Runner
}

func CLI() *CommandLineInterface {

	cli := CommandLineInterface{}
	cli.LoadConfiguration()

	return &cli
}

func (cli *CommandLineInterface) Build(args []string) (Runner, error) {

	if len(args) == 0 || args[0] == "" {
		return nil, errors.New("Missing required arguments; run with 'help' to see usage")
	}

	log.Println("[DBG] args:", args)

	switch operation := strings.ToLower(args[0]); operation {
	case SERVER:
		{
			log.Println("Server is running...")

			application, err := NewApplication()

			if err != nil {
				log.Fatal(err)
				os.Exit(0)
			}
			return application, nil
		}
	case MONITOR:
		{
			log.Println("Incoming monitor command with parameters", args[1:])
		}
	default:
		{
			return nil, errors.New("Invalid input operation '" + operation + "' exception")
		}
	}

	return nil, errors.New("Not implemented exception")
}

func (cli CommandLineInterface) LoadConfiguration() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}
