package main

import (
	"log"

	"github.com/gabriel-barreto/go-quoting-api/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error while trying to load .env file: %s", err)
	}
	server.Perform()
}
