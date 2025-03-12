package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/joho/godotenv"
)

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

func getUserID(client *twitter.Client, username string) (string, error) {
	opts := twitter.UserLookupOpts{
		UserFields: []twitter.UserField{
			twitter.UserFieldID,
			twitter.UserFieldName,
		},
	}

	users, err := client.UserNameLookup(context.Background(), []string{username}, opts)
	if err != nil {
		return "", fmt.Errorf("failed to lookup user: %v", err)
	}

	if len(users.Raw.Users) == 0 {
		return "", fmt.Errorf("no user found with username: %s", username)
	}

	return users.Raw.Users[0].ID, nil
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get bearer token from environment
	token := os.Getenv("X_V2_BEARER_TOKEN")
	if token == "" {
		log.Fatal("Bearer token not found in environment")
	}

	// Create client with proper HTTP client initialization
	client := &twitter.Client{
		Authorizer: authorize{Token: token},
		Client:     &http.Client{},
		Host:       "https://api.twitter.com",
	}

	// Get user ID from environment or lookup from username
	var userID string
	if envUserID := os.Getenv("X_USERID"); envUserID != "" {
		userID = envUserID
		log.Printf("Using user ID from environment: %s\n", userID)
	} else {
		username := os.Getenv("X_USERNAME")
		if username == "" {
			log.Fatal("Neither X_USERID nor X_USERNAME found in environment")
		}
		var err error
		userID, err = getUserID(client, username)
		if err != nil {
			log.Fatalf("Failed to get user ID: %v", err)
		}
		log.Printf("Found user ID: %s for username: %s\n", userID, username)
	}

	// Calculate cutoff date
	cutoffYears := 1 // default value
	if cutoffYearsStr := os.Getenv("X_CUTOFF_YEARS"); cutoffYearsStr != "" {
		if years, err := strconv.Atoi(cutoffYearsStr); err == nil {
			cutoffYears = years
		}
	}
	cutoffTime := time.Now().AddDate(cutoffYears, 0, 0)
	fmt.Printf("Fetching tweets older than: %s\n", cutoffTime.Format("2006-01-02"))

	maxResults := 5 // default value
	if limitStr := os.Getenv("X_LIMIT_TWEET_FETCH"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			maxResults = limit
		}
	}

	// Get tweets with proper options
	opts := twitter.UserTweetTimelineOpts{
		MaxResults: maxResults,
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldText,
		},
	}

	// Fetch user tweets using the correct method
	tweets, err := client.UserTweetTimeline(context.Background(), userID, opts)
	if err != nil {
		log.Fatalf("Failed to fetch tweets: %v", err)
	}

	// Process tweets
	for _, tweet := range tweets.Raw.Tweets {
		createdAt, err := time.Parse(time.RFC3339, tweet.CreatedAt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing time for tweet %s: %v\n", tweet.ID, err)
			continue
		}

		if createdAt.Before(cutoffTime) {
			fmt.Printf("Would delete tweet ID %s: %s\n", tweet.ID, tweet.Text)
			// Uncomment to actually delete tweets
			// _, err := client.DeleteTweet(context.Background(), tweet.ID)
			// if err != nil {
			//     fmt.Fprintf(os.Stderr, "Error deleting tweet %s: %v\n", tweet.ID, err)
			// } else {
			//     fmt.Printf("Deleted tweet %s\n", tweet.ID)
			// }
		}
	}
}
