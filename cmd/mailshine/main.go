package main

import (
	"flag"
	"log"
	"os"

	"github.com/hebo/mailshine/models"
	"github.com/hebo/mailshine/providers"
	"github.com/hebo/mailshine/server"
	"github.com/hebo/mailshine/service"
	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml"
)

var (
	configFilename = "config.toml"
	dbPath         = "./shine.db"
)

func main() {
	flagGenFeeds := flag.Bool("generate", false, "Generate feeds")
	flagInit := flag.Bool("init", false, "Initialize database")
	flagPort := flag.Int("port", 8080, "Listen port")
	flag.Parse()

	conf := loadConfig()
	log.Printf("Config loaded - %d feed configs found\n", len(conf.FeedConfigs))

	db, err := models.NewDB(dbPath)
	if err != nil {
		log.Fatalln("could not get db", err)
	}

	if *flagInit {
		err := db.InitializeSchema()
		if err != nil {
			log.Println("Error initializing database", err)
		}
		return
	}

	reddit, err := providers.NewRedditClient(
		os.Getenv("REDDIT_CLIENT_ID"),
		os.Getenv("REDDIT_CLIENT_SECRET"))
	if err != nil {
		log.Fatalf("Failed to create reddit client: %s", err)
	}

	svc := service.NewService(db, conf.FeedConfigs, reddit)
	svc.StartScheduler()

	if *flagGenFeeds {
		svc.CreateAllDigests()
		return
	}

	server := server.NewServer(db, conf.FeedConfigs, conf.BaseURL)
	server.Serve(*flagPort)
}

type config struct {
	BaseURL     string               `toml:"base_url"`
	FeedConfigs models.FeedConfigMap `toml:"feeds"`
}

func loadConfig() config {
	fc := config{}
	fi, err := os.Open(configFilename)
	if err != nil {
		log.Fatalf("error opening file %s\n", err)
	}
	err = toml.NewDecoder(fi).Decode(&fc)
	if err != nil {
		log.Fatalf("error decoding %s\n", err)
	}

	for _, f := range fc.FeedConfigs {
		err = f.Validate()
		if err != nil {
			log.Fatalf("Invalid config: %s\n", err)
		}
	}

	err = godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load '.env' file")
	}

	if os.Getenv("DB_PATH") != "" {
		dbPath = os.Getenv("DB_PATH")
	}

	return fc
}
