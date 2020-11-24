package service

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hebo/mailshine/models"
	"github.com/hebo/mailshine/providers"
)

type Service struct {
	db           models.DB
	feeds        models.FeedConfigMap
	redditClient providers.RedditClient
}

// NewService creates a new Service
func NewService(db models.DB, fc models.FeedConfigMap, reddit providers.RedditClient) Service {
	svc := Service{
		db:           db,
		feeds:        fc,
		redditClient: reddit,
	}

	return svc
}

const timezone = "America/Los_Angeles"

// StartScheduler begins scheduling for Digest generation
func (s Service) StartScheduler() {

	loc, err := time.LoadLocation(timezone) // use other time zones such as MST, IST
	if err != nil {
		log.Fatalln("failed to get timezone: ", err)
	}

	scheduler := gocron.NewScheduler(loc)

	for name := range s.feeds {
		count, err := s.db.CountDigestsByFeed(name)
		if err != nil {
			log.Printf("Failed to get digest count: %s", err)
			continue
		}

		if count == 0 {
			log.Printf("No digests for feed %q, fetching initial\n", name)
			s.createDigest(name)
		}

		scheduler.Every(1).Day().At("8:00").Do(func() {
			log.Printf("Scheduler triggered for %q\n", name)
			s.createDigest(name)
		})
	}

}

// createDigest generates and stores a new digest
func (s Service) createDigest(feedName string) error {
	log.Printf("Processing feed %q", feedName)
	feedConf := s.feeds[feedName]

	count, err := s.db.CountDigestsByFeed(feedName)
	if err != nil {
		log.Printf("Failed to get digest count: %s", err)
		return err
	}

	dg := models.Digest{
		Title:     fmt.Sprintf("%s #%d", feedConf.Title, count+1),
		FeedName:  feedName,
		CreatedAt: time.Now(),
	}

	for _, subreddit := range feedConf.Reddits {
		listing, err := s.redditClient.FetchSubreddit(subreddit, feedConf.NumItems)
		if err != nil {
			return fmt.Errorf("fetch subreddit %q: %w", subreddit, err)
		}

		blk := listing.ToBlock("r/" + subreddit)
		dg.Content = append(dg.Content, blk)
	}

	err = s.db.InsertDigest(dg)
	if err != nil {
		return fmt.Errorf("failed to insert feed: %s", err)
	}
	log.Printf("Inserted feed: %q\n", feedName)
	return nil
}

// CreateAllDigests generates a digest for all configured feeds
func (s Service) CreateAllDigests() error {
	for name := range s.feeds {
		err := s.createDigest(name)
		if err != nil {
			return err
		}
	}
	return nil
}