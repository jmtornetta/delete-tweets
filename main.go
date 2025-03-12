package main

import (
    "fmt"
    "log"
    "os"
    "time"

    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    // Get environment variables
    apiKey := os.Getenv("X_API_KEY")
    apiSecretKey := os.Getenv("X_API_SECRET_KEY")
    accessToken := os.Getenv("X_ACCESS_TOKEN")
    accessTokenSecret := os.Getenv("X_ACCESS_TOKEN_SECRET")

    // Authenticate using OAuth1
    config := oauth1.NewConfig(apiKey, apiSecretKey)
    token := oauth1.NewToken(accessToken, accessTokenSecret)
    httpClient := config.Client(oauth1.NoContext, token)

    client := twitter.NewClient(httpClient)

    // Calculate cutoff date (1 year ago)
    oneYearAgo := time.Now().AddDate(-1, 0, 0)

    fmt.Println("Fetching tweets older than:", oneYearAgo.Format("2006-01-02"))

    // Get tweets
    tweets, resp, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
        Count:     200, // Max 200 tweets per request
        TweetMode: "extended",
    })
    if err != nil || resp.StatusCode != 200 {
        log.Fatalf("Failed to fetch tweets: %v", err)
    }

    // Delete tweets older than 1 year
    for _, tweet := range tweets {
        tweetTime, _ := tweet.CreatedAtTime()
        if tweetTime.Before(oneYearAgo) {
            fmt.Printf("Deleting tweet ID %d: %s\n", tweet.ID, tweet.FullText)
            // _, _, err := client.Statuses.Destroy(tweet.ID, nil) // Uncomment to delete
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error deleting tweet %d: %v\n", tweet.ID, err)
            } else {
                fmt.Println("Deleted successfully.")
            }
        }
    }
}