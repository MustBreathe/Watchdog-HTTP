package config

import (
	"log"
	"os"
	"slices"
	"strings"
)

type DisplayConfiguration struct {
	// Display configuration for the watchdog application
	// This struct can be extended to include more display-related settings as needed
}

func NewDisplayConfiguration() (*DisplayConfiguration, error) {
	return &DisplayConfiguration{}, nil
}

func (dc *DisplayConfiguration) Run() error {
	log.Println("Current configuration loaded")
	conf := os.Environ()
	slices.Sort(conf)

	for _, e := range conf {
		if strings.HasPrefix(e, "WATCHDOG_") {
			log.Println(e)
		}
	}

	for _, e := range conf {
		if !strings.HasPrefix(e, "WATCHDOG_") {
			log.Println(e)
		}
	}
	return nil
}
