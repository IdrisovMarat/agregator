package command

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/IdrisovMarat/agregator/internal/config"
	"github.com/IdrisovMarat/agregator/internal/database"
	"github.com/IdrisovMarat/agregator/internal/xml"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
	// const url = "https://www.wagslane.dev/index.xml"

	if len(cmd.Args) != 1 {
		fmt.Fprintln(os.Stderr, "agg command requires the time_between_reqs")
		os.Exit(1)
	}

	timeBetweenReqs := cmd.Args[0]

	// Парсим duration
	duration, err := time.ParseDuration(timeBetweenReqs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing duration '%s': %v", timeBetweenReqs, err)
	}

	// Проверяем, что duration не слишком маленький (анти-DOS защита)
	if duration < time.Second*10 {
		fmt.Fprintf(os.Stderr, "Duration too short: %v. Minimum is 10 seconds to avoid overloading servers.", duration)
	}

	fmt.Printf("🚀 Starting RSS aggregator\n")
	fmt.Printf("⏰ Collecting feeds every %v\n", duration)
	fmt.Printf("⏹️  Press Ctrl+C to stop\n\n")

	// Обработка сигналов для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Запускаем ticker
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	// Выполняем немедленно первую агрегацию
	fmt.Println("🔄 Starting initial fetch...")
	if err := ScrapeFeeds(ctx, s); err != nil {
		log.Printf("Error in initial fetch: %v", err)
	}

	// Основной цикл
	for {
		select {
		case <-ticker.C:
			if err := ScrapeFeeds(ctx, s); err != nil {
				fmt.Printf("Error scraping feeds: %v", err)
			}

		case <-sigCh:
			fmt.Println("\n🛑 Shutting down aggregator...")
			return nil
		}

	}

	// return nil
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

// handlerFollow handles the follow command
func HandlerUnFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("unfollow command requires an url")
	}

	ctx2 := context.Background()

	feedId, err := s.Db.GetFeedIdByUrl(ctx2, cmd.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to : %v", err)
		os.Exit(1)
	}

	param := database.DeleteFeedFollowByUserAndFeedParams{
		UserID: user.ID,
		FeedID: feedId,
	}

	ctx3 := context.Background()

	err = s.Db.DeleteFeedFollowByUserAndFeed(ctx3, param)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find : %v", err)
		os.Exit(1)
	}
	fmt.Print("\n****************************************************************\n")
	fmt.Println("DELETED")
	fmt.Print("\n****************************************************************\n")
	return nil
}

func ScrapeFeeds(ctx context.Context, s *State) error {

	feed, err := s.Db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("error getting next feed to fetch: %w", err)
	}

	err = s.Db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %w", err)
	}

	log.Printf("📡 Fetching feed: %s (%s)", feed.Name, feed.Url)

	// Получаем RSS фид
	rssFeed, err := xml.FetchFeed(ctx, feed.Url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching feed %s: %v", feed.Url, err)
		os.Exit(1)
	}

	// Сохраняем посты в базу данных вместо вывода в консоль
	postsSaved := 0
	for _, item := range rssFeed.Channel.Item {
		err := savePost(ctx, s, feed.ID, item)
		if err != nil {
			log.Printf("⚠️ Error saving post '%s': %v", item.Title, err)
			continue
		}
		postsSaved++
	}

	log.Printf("✅ Saved %d new posts from %s", postsSaved, feed.Name)
	return nil
}

func savePost(ctx context.Context, s *State, feedID uuid.UUID, item xml.RSSItem) error {
	// Парсим дату публикации
	publishedAt, err := parsePubDate(item.PubDate)
	if err != nil {
		log.Printf("⚠️ Could not parse date '%s': %v", item.PubDate, err)
		// Используем текущее время если дата невалидна
		publishedAt = time.Now()
	}

	// Очищаем и обрезаем слишком длинные поля
	title := cleanText(item.Title, 500)
	description := cleanText(item.Description, 2000)
	url := cleanText(item.Link, 500)

	savePostParam := database.CreatePostParams{
		Title:       title,
		Url:         url,
		Description: sql.NullString{String: description, Valid: true},
		PublishedAt: sql.NullTime{Time: publishedAt, Valid: true},
		FeedID:      feedID,
	}

	// Пытаемся создать пост
	_, err = s.Db.CreatePost(ctx, savePostParam)

	// Игнорируем ошибку дубликата URL
	if err != nil {
		if isDuplicateError(err) {
			return nil // Просто игнорируем дубликаты
		}
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

// parsePubDate парсит различные форматы дат из RSS
func parsePubDate(dateStr string) (time.Time, error) {
	// Убираем лишние пробелы
	dateStr = strings.TrimSpace(dateStr)

	// Попробуем несколько распространенных форматов RSS
	formats := []string{
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC822,   // "02 Jan 06 15:04 MST"
		time.RFC822Z,  // "02 Jan 06 15:04 -0700"
		time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 MST",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// cleanText очищает текст и обрезает до максимальной длины
func cleanText(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if len(text) > maxLen {
		return text[:maxLen]
	}
	return text
}

// isDuplicateError проверяет, является ли ошибка ошибкой дубликата
func isDuplicateError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return strings.Contains(err.Error(), "duplicate") ||
		strings.Contains(err.Error(), "23505")
}

// formatDate форматирует дату для красивого вывода
func formatDate(t time.Time) string {
	now := time.Now()

	// Если сегодня
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		return "Today at " + t.Format("15:04")
	}

	// Если вчера
	yesterday := now.AddDate(0, 0, -1)
	if t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day() {
		return "Yesterday at " + t.Format("15:04")
	}

	// Если в этом году
	if t.Year() == now.Year() {
		return t.Format("Jan 2 at 15:04")
	}

	// Если в другом году
	return t.Format("Jan 2, 2006")
}

// truncateText обрезает текст и добавляет многоточие
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// HandlerBrowse получает посты для пользователя и красиво выводит
func HandlerBrowse(s *State, cmd Command, user database.User) error {

	// Парсим лимит (по умолчанию 2)
	limit := 2
	if len(cmd.Args) > 0 {
		parsedLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil || parsedLimit <= 0 {
			log.Fatalf("Invalid limit: %s. Please provide a positive number.", cmd.Args[0])
		}
		limit = parsedLimit
	}
	ctx := context.Background()
	// Получаем посты для пользователя
	posts, err := s.Db.GetPostsForUser(ctx, database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		log.Fatalf("Error getting posts: %v", err)
	}

	if len(posts) == 0 {
		fmt.Println("📭 No posts found.")
		fmt.Println("   Make sure you're following some feeds and that the aggregator is running!")
		return nil
	}

	fmt.Printf("📰 Latest %d posts from your feeds:\n\n", len(posts))

	for i, post := range posts {
		fmt.Printf("┌─── Post %d ──────────────────────────────────────────\n", i+1)
		fmt.Printf("│ 📝 %s\n", post.Title)
		fmt.Printf("│ 📅 Published: %s\n", formatDate(post.PublishedAt.Time))
		fmt.Printf("│ 📋 Feed: %s\n", post.FeedName)
		fmt.Printf("│ 🔗 %s\n", post.Url)

		if post.Description.String != "" {
			// Обрезаем и форматируем описание
			desc := truncateText(post.Description.String, 200)
			fmt.Printf("│ 📄 %s\n", desc)
		}

		fmt.Printf("└─────────────────────────────────────────────────────\n\n")
	}
	return nil
}
