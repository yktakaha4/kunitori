package pkg

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

const GitHubAccessTokenKey = "GITHUB_ACCESS_TOKEN"

func ShowSlowMessage() {
	if !IsGitHubAccessTokenProvided() {
		println(fmt.Sprintf(
			"warning: If the environment variable %v is not set, API searches will be very slow.",
			GitHubAccessTokenKey,
		))
	}
}

func FindLoginByEmail(email string) (string, error) {
	// https://docs.github.com/ja/rest/search?apiVersion=2022-11-28#search-users
	sleep := 7
	if IsGitHubAccessTokenProvided() {
		sleep = 3
	}

	if email == "" {
		return "", nil
	}

	log.Printf("start FindLoginByEmail: email=%v, sleep=%v", email, sleep)

	gitHubNoReplyEmailRegex := regexp.MustCompile("@users\\.noreply\\.github\\.com$")

	if gitHubNoReplyEmailRegex.MatchString(email) {
		emailHost := gitHubNoReplyEmailRegex.ReplaceAllString(email, "")
		parts := strings.Split(emailHost, "+")
		if len(parts) == 2 {
			login := parts[1]
			log.Printf("github user email: email=%v, login=%v", email, login)

			return login, nil
		} else {
			return "", fmt.Errorf("invalid github email: email=%v", email)
		}
	}

	query := fmt.Sprintf("%v in:email", email)

	log.Printf("search users: query=%v", query)

	client, ctx := createGitHubClient()

	result, _, err := client.Search.Users(ctx, query, nil)

	time.Sleep(time.Duration(sleep) * time.Second)

	if err != nil {
		return "", err
	}

	if result.GetTotal() > 0 {
		login := *result.Users[0].Login

		log.Printf("user found: email=%v, login=%v", email, login)
		return login, nil
	} else {
		log.Printf("user not found: email=%v", email)
		return "", nil
	}
}

func IsGitHubAccessTokenProvided() bool {
	return len(os.Getenv(GitHubAccessTokenKey)) > 0
}

func createGitHubClient() (*github.Client, context.Context) {
	ctx := context.Background()
	if IsGitHubAccessTokenProvided() {
		token := os.Getenv(GitHubAccessTokenKey)

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)

		return github.NewClient(tc), ctx
	} else {
		return github.NewClient(nil), ctx
	}
}
