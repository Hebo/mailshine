package models

import "fmt"

type FeedConfigMap map[string]FeedConfig

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

var validTimePeriods = []string{"day", "week"}

// Validate performs basic checks to ensure FeedConfig is initialized properly
func (c FeedConfig) Validate() error {
	if c.NumItems == 0 {
		return fmt.Errorf("feed Config %q: NumItems is not set", c.Title)
	}

	if c.Schedule == "" {
		return fmt.Errorf("feed Config %q: Schedule is not set", c.Title)
	}

	if !contains(validTimePeriods, c.TimePeriod) {
		return fmt.Errorf("feed Config %q: invalid TimePeriod", c.Title)
	}

	return nil
}

// FeedConfig is the configuration for a single Feed
type FeedConfig struct {
	Title      string
	Reddits    []string
	NumItems   int    `toml:"num_items"`
	TimePeriod string `toml:"time_period"`
	// Schedule is in crontab syntax
	Schedule string `toml:"schedule"`
}
