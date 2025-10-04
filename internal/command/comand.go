package command

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IdrisovMarat/agregator/internal/config"
	"github.com/IdrisovMarat/agregator/internal/database"
	"github.com/IdrisovMarat/agregator/internal/xml"
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

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {

	return func(s *State, cmd Command) error {
		usersName := s.Config.CurrentUserName

		ctx1 := context.Background()

		user, err := s.Db.GetUser(ctx1, usersName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to find user in s.Db.GetUser() function - the user does not exists: %v", err)
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}

// run executes a command with the given state
func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}

// registerLoggedIn регистрирует обычную команду
func (c *Commands) Register(name string, handler func(*State, Command) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*State, Command) error)
	}
	c.handlers[name] = handler
}

// registerLoggedIn регистрирует команду, требующую авторизации
func (c *Commands) RegisterLoggedIn(name string, handler func(*State, Command, database.User) error) {

	if c.handlers == nil {
		c.handlers = make(map[string]func(*State, Command) error)
	}

	c.handlers[name] = middlewareLoggedIn(handler)
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

// handlerAgg handles the aggregator service
func HandlerAgg(s *State, cmd Command) error {
	const url = "https://www.wagslane.dev/index.xml"
	ctx := context.Background()

	rssFeed, err := xml.FetchFeed(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to reset the table 'users': %v", err)
		os.Exit(1)
	}

	fmt.Println(rssFeed.Channel.Title)
	fmt.Println(rssFeed.Channel.Description)
	fmt.Println("******************************************")
	for i := range rssFeed.Channel.Item {
		fmt.Println(rssFeed.Channel.Item[i].Title)
		fmt.Println("---------------------------------------------------------------------------------------------------------")
		fmt.Println(rssFeed.Channel.Item[i].Description)
		fmt.Println("**********************************************************************************************************")
	}
	return nil
}

// handlerAddfeed handles the addfeed command for adding feeds
func HandlerAddfeed(s *State, cmd Command, user database.User) error {

	if len(cmd.Args) != 2 {
		fmt.Fprintln(os.Stderr, "addfeed command requires the url and name")
		os.Exit(1)
	}

	feedParam := database.CreateFeedParams{
		Name:   cmd.Args[0],
		Url:    cmd.Args[1],
		UserID: user.ID,
	}

	ctx1 := context.Background()

	feed, err := s.Db.CreateFeed(ctx1, feedParam)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create feed through createfeed: %v", err)
	}

	folowParam := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	ctx3 := context.Background()

	followItems, err := s.Db.CreateFeedFollow(ctx3, folowParam)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find : %v", err)
		os.Exit(1)
	}

	fmt.Print("\n****************************************************************\n")
	fmt.Printf("%v\n%v\n%v\n%v\n", followItems.UserName, followItems.FeedName, feed.Url, followItems.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Print("\n****************************************************************\n")

	return nil
}

// handlerFeeds handles the feeds command
func HandlerFeeds(s *State, cmd Command) error {

	ctx := context.Background()

	feeds, err := s.Db.GetFeedsWithName(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get feeds from the table 'feeds': %v", err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		fmt.Printf("%v\n%v\n%v\n%v\n", feed.Name_2, feed.Name.String, feed.Url.String, feed.CreatedAt.Time.Format("2006-01-02 15:04:05"))
		fmt.Print("\n****************************************************************\n")
	}

	return nil
}

// handlerFollow handles the follow command
func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("follow command requires an url")
	}

	ctx2 := context.Background()

	feedId, err := s.Db.GetFeedIdByUrl(ctx2, cmd.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to : %v", err)
		os.Exit(1)
	}

	param := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedId,
	}

	ctx3 := context.Background()

	followItems, err := s.Db.CreateFeedFollow(ctx3, param)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find : %v", err)
		os.Exit(1)
	}
	fmt.Print("\n****************************************************************\n")
	fmt.Printf("%v\n%v\n%v\n", followItems.UserName, followItems.FeedName, followItems.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Print("\n****************************************************************\n")
	return nil
}

// handlerFollowing handles the following command
func HandlerFollowing(s *State, cmd Command, user database.User) error {

	ctx2 := context.Background()

	feeds, err := s.Db.GetFeedFollowsForUser(ctx2, user.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to : %v", err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		fmt.Print("\n****************************************************************\n")
		fmt.Printf("\n\n%v\n%v\n%v\n\n", feed.UserName, feed.FeedName, feed.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Print("\n****************************************************************\n")
	}

	return nil
}
