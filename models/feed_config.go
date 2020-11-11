package models

import "fmt"

type FeedConfigMap map[string]FeedConfig

func (c FeedConfig) Validate() error {
	if c.NumItems == 0 {
		return fmt.Errorf("Feed Config %q: NumItems is not set", c.Title)
	}
	return nil
}

type FeedConfig struct {
	Title    string
	Reddits  []string
	NumItems int `toml:"num_items"`
}
