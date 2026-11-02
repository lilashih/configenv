package main

import (
	"fmt"
	"log"

	"github.com/lilashih/configenv/example/config"
)

func main() {
	if err := config.Initialize(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Settings loaded with Load:")
	printConfig()

	result, err := config.InitializeWithResult()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n\nSettings loaded with LoadWithResult:")

	if result.EnvFile == "" {
		fmt.Println("No .env file found, using process environment variables and defaults")
	} else {
		fmt.Printf("Config file: %s\n", result.EnvFile)
	}
	printConfig()
}

func printConfig() {
	fmt.Printf("Version: %s\n", config.App.Version)
	fmt.Printf("Mode: %s\n", config.App.Mode)
	fmt.Printf("Port: %s\n", config.App.Port)
	fmt.Printf("Database path: %s", config.DB.Path)
}
