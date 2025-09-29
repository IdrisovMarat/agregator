package command

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IdrisovMarat/agregator/internal/config"
	"github.com/IdrisovMarat/agregator/internal/database"
	"github.com/google/uuid"
)

// State holds the application state
type State struct {
	Db     *database.Queries
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

	usersName := cmd.Args[0]

	ctx := context.Background()

	user, err := s.Db.GetUser(ctx, usersName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find user in s.Db.GetUser() function - the user does not exists: %v", err)
		os.Exit(1)
	}

	if err := s.Config.SetUser(user.Name); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Printf("User set to: %s\n", user.Name)
	return nil
}

// handlerRegister handles the register command
func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("register command requires a username")
	}

	argUser := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}

	ctx := context.Background()

	user, err := s.Db.CreateUser(ctx, argUser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create user in s.Db.CreateUser() function - the user already exists: %v", err)
		os.Exit(1)
	}

	if err := s.Config.SetUser(user.Name); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Printf("The user was successfully created: %s\n", user)
	return nil
}

// handlerReset handles the register command
func HandlerReset(s *State, cmd Command) error {

	ctx := context.Background()

	err := s.Db.ResetTable(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to reset the table 'users': %v", err)
		os.Exit(1)
	}

	fmt.Println("The user table was successfully deleted")
	return nil
}

// handlerReset handles the register command
func HandlerGetUsers(s *State, cmd Command) error {

	ctx := context.Background()

	users, err := s.Db.GetUsers(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get users from the table 'users': %v", err)
		os.Exit(1)
	}
	var current string = ""
	for _, user := range users {
		if s.Config.CurrentUserName == user {
			current = "(current)"
		}
		fmt.Printf("* %v %v\n", user, current)
	}

	return nil
}
