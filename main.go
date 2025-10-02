package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error connecting to database: %v\n", err)
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
	commands.Register("users", command.HandlerGetUsers)
	commands.Register("agg", command.HandlerAgg)
	commands.Register("addfeed", command.HandlerAddfeed)
	commands.Register("feeds", command.HandlerFeeds)
	commands.Register("follow", command.HandlerFollow)
	commands.Register("following", command.HandlerFollowing)

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
