package command

import (
	"fmt"

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
func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}

// register adds a new command handler
func (c *Commands) Register(name string, handler func(*State, Command) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*State, Command) error)
	}
	c.handlers[name] = handler
}

// handlerLogin handles the login command
func HandlerLogin(s *State, cmd Command) error {
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
