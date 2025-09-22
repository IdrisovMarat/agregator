package main

import (
	"fmt"
	"log"

	"github.com/IdrisovMarat/agregator/internal/config"
)

func main() {
	// Read the config file
	fmt.Println("Reading config file...")
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	if cfg.CurrentUserName != "" {
		fmt.Printf("Initial config: DBURL=%s, CurrentUser=%s\n", cfg.DBURL, cfg.CurrentUserName)
	} else {
		fmt.Printf("Initial config: DBURL=%s\n", cfg.DBURL)
		fmt.Println("Type your user name, please...")
	}

	// Set the current user to your name
	var yourName string
	fmt.Scanln(&yourName) // Читает до пробела/перевода строки
	fmt.Printf("Setting current user to: %s\n", yourName)
	if err := cfg.SetUser(yourName); err != nil {
		log.Fatalf("Failed to set user: %v", err)
	}

	// Read the config file again to verify the changes
	fmt.Println("Reading config file again to verify changes...")
	updatedCfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read updated config: %v", err)
	}

	// Print the contents of the config struct
	fmt.Println("Updated config:")
	fmt.Printf("  DBURL: %s\n", updatedCfg.DBURL)
	fmt.Printf("  CurrentUserName: %s\n", updatedCfg.CurrentUserName)
}
