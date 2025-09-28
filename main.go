package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/IdrisovMarat/agregator/internal/command"
	"github.com/IdrisovMarat/agregator/internal/config"
	"github.com/IdrisovMarat/agregator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	// Read the config file
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	dbURL := cfg.DBURL

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	dbQueries := database.New(db)

	// Create application state
	state := &command.State{
		Db:     dbQueries,
		Config: cfg,
	}

	// Initialize commands registry
	commands := &command.Commands{}
	commands.Register("login", command.HandlerLogin)
	commands.Register("register", command.HandlerRegister)
	commands.Register("reset", command.HandlerReset)

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: not enough arguments provided")
		fmt.Fprintln(os.Stderr, "Usage: gator <command> [arguments]")
		os.Exit(1)
	}

	// Create command from arguments
	cmd := command.Command{
		Name: os.Args[1],
		Args: []string{},
	}

	// Add remaining arguments if they exist
	if len(os.Args) > 2 {
		cmd.Args = os.Args[2:]
	}

	// Execute the command
	if err := commands.Run(state, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
