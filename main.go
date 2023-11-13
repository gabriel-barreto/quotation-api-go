package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/gabriel-barreto/go-quoting-api/client"
	"github.com/gabriel-barreto/go-quoting-api/server"
	"github.com/joho/godotenv"
)

func findArgTarget(args []string) (string, error) {
	for _, arg := range args {
		if strings.Contains(arg, "--") {
			return arg, nil
		}
	}
	return "", errors.New("unknown target")
}

func main() {
	target, err := findArgTarget(os.Args)
	if err != nil {
		log.Fatalf("Error while trying to get command target: %s", err)
	}
	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error while trying to load .env file: %s", err)
	}
	if target == "--server" {
		server.Start()
	}
	if target == "--client" {
		client.Perform()
	}
}
