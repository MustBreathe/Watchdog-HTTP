package cli

import (
	"cmd/watchdog/main.go/internal/app"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const SERVER = "server"
const MONITOR = "monitor"

type IExecutable interface {
	Execute()
}

type CommandLineInterface struct {
	strategy IExecutable
}

func CLI() *CommandLineInterface {

	cli := CommandLineInterface{}
	cli.LoadConfiguration()

	return &cli
}

func (cli *CommandLineInterface) Execute(args []string) error {

	if len(args) == 0 || args[0] == "" {
		return errors.New("Missing required arguments; run with 'help' to see usage")
	}

	switch operation := strings.ToLower(args[0]); operation {
	case SERVER:
		{
			log.Println("Build Server strategy: ", args)
			//cli.strategy = ServerStrategy{}

			application, err := app.New()

			if err != nil {
				log.Fatal(err)
				os.Exit(0)
			}
			application.Run()
		}
	case MONITOR:
		{
			log.Println("Build Monitor strategy: ", args)
		}
	default:
		{
			return errors.New("Invalid input operation '" + operation + "' exception")
		}
	}

	return nil
}

func (cli CommandLineInterface) LoadConfiguration() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}
