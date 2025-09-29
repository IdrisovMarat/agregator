package xml

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// url := "https://www.wagslane.dev/index.xml"

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	var feed RSSFeed

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("req formation with NewRequestWithContext error: %s", err)
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", "agregator")

	resp, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("client.Do req error: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &RSSFeed{}, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(&feed)
	// xmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("xml decoding error: %s", err)
	}
	// xmlstr := html.UnescapeString()
	// err = xml.Unmarshal(xmlBytes, &rssFeed)
	// if err != nil {
	// 	return &RSSFeed{}, fmt.Errorf("xml decoding error: %s", err)
	// }
	// Обрабатываем HTML entities в полях
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}
