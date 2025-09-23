package main

import (
	"fmt"
	"log"
	"os"

	"github.com/IdrisovMarat/agregator/internal/config"
)

// State holds the application state
type State struct {
	Config *config.Config
}

// Command represents a CLI command
type Command struct {
	Name string
	Args []string
}

// Commands holds all registered command handlers
type Commands struct {
	handlers map[string]func(*State, Command) error
}

// run executes a command with the given state
func (c *Commands) run(s *State, cmd Command) error {
	handler, exists := c.handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}

// register adds a new command handler
func (c *Commands) register(name string, handler func(*State, Command) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*State, Command) error)
	}
	c.handlers[name] = handler
}

// handlerLogin handles the login command
func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("login command requires a username")
	}

	username := cmd.Args[0]
	if err := s.Config.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Printf("User set to: %s\n", username)
	return nil
}

func main() {
	// Read the config file
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// Create application state
	state := &State{Config: cfg}

	// Initialize commands registry
	commands := &Commands{}
	commands.register("login", handlerLogin)

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: not enough arguments provided")
		fmt.Fprintln(os.Stderr, "Usage: gator <command> [arguments]")
		os.Exit(1)
	}

	// Create command from arguments
	cmd := Command{
		Name: os.Args[1],
		Args: []string{},
	}

	// Add remaining arguments if they exist
	if len(os.Args) > 2 {
		cmd.Args = os.Args[2:]
	}

	// Execute the command
	if err := commands.run(state, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// func main() {
// 	// // Read the config file
// 	// fmt.Println("Reading config file...")
// 	// cfg, err := config.Read()
// 	// if err != nil {
// 	// 	log.Fatalf("Failed to read config: %v", err)
// 	// }

// 	// if cfg.CurrentUserName != "" {
// 	// 	fmt.Printf("Initial config: DBURL=%s, CurrentUser=%s\n", cfg.DBURL, cfg.CurrentUserName)
// 	// } else {
// 	// 	fmt.Printf("Initial config: DBURL=%s\n", cfg.DBURL)
// 	// 	fmt.Println("Type your user name, please...")
// 	// }

// 	// // Set the current user to your name
// 	// var yourName string
// 	// fmt.Scanln(&yourName) // Читает до пробела/перевода строки
// 	// fmt.Printf("Setting current user to: %s\n", yourName)
// 	// if err := cfg.SetUser(yourName); err != nil {
// 	// 	log.Fatalf("Failed to set user: %v", err)
// 	// }

// 	// // Read the config file again to verify the changes
// 	// fmt.Println("Reading config file again to verify changes...")
// 	// updatedCfg, err := config.Read()
// 	// if err != nil {
// 	// 	log.Fatalf("Failed to read updated config: %v", err)
// 	// }

// 	// // Print the contents of the config struct
// 	// fmt.Println("Updated config:")
// 	// fmt.Printf("  DBURL: %s\n", updatedCfg.DBURL)
// 	// fmt.Printf("  CurrentUserName: %s\n", updatedCfg.CurrentUserName)
// }
