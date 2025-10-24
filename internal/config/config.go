// Package config contains the logic to retrieve application specific environment configurations
package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

var (
	// ErrMissingVariable is returned by GetConfig if some of the required variables are missing.
	ErrMissingVariable = errors.New("required variable is missing")
	// ErrMissingPrefix is returned by GetConfig if some of the variables prefix is incorrect.
	ErrMissingPrefix = errors.New("mandatory prefix is missing")
)

// InDebugMode determines if the application is running in debug mode base.
//
// Returns true if the environment variable `DEBUG` has a value of either "1", "true" or "enable", false in every other case.
func InDebugMode() bool {
	debugEnabledOptions := []string{"1", "true", "enable"}

	return slices.Contains(debugEnabledOptions, strings.ToLower(os.Getenv("DEBUG")))
}

// GetConfig parses the Slack Bot's required credentials from the environment.
//
// return the bot token, app token and an error if any.
func GetConfig() (string, string, error) {
	var (
		botToken = os.Getenv("SLACK_BOT_TOKEN")
		appToken = os.Getenv("SLACK_APP_TOKEN")
	)

	if botToken == "" {
		return "", "", fmt.Errorf("SLACK_BOT_TOKEN: %w", ErrMissingVariable)
	}

	if appToken == "" {
		return "", "", fmt.Errorf("SLACK_APP_TOKEN: %w", ErrMissingVariable)
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		return "", "", fmt.Errorf("SLACK_BOT_TOKEN: %w, prefix: xoxb-", ErrMissingPrefix)
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		return "", "", fmt.Errorf("SLACK_APP_TOKEN: %w, prefix: xapp-", ErrMissingPrefix)
	}

	return botToken, appToken, nil
}
