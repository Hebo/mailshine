package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/feeds"
	"github.com/gorilla/handlers"
	"github.com/hebo/mailshine/models"
	"github.com/julienschmidt/httprouter"
)

// Server handles rendering RSS and info for feeds
type Server struct {
	Feeds   models.FeedConfigMap
	router  *httprouter.Router
	db      models.DB
	baseURL string
}

// NewServer creates a new server
func NewServer(db models.DB, feeds models.FeedConfigMap, baseURL string) Server {
	srv := Server{
		Feeds:   feeds,
		db:      db,
		baseURL: baseURL,
	}

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/feeds/:name", srv.GetFeed)
	router.GET("/feeds/:name/rss", srv.GetFeedRSS)
	router.GET("/feeds/:name/preview", srv.GetPreviewRSS)

	router.ServeFiles("/static/*filepath", http.Dir("static"))

	srv.router = router
	return srv
}

// Serve starts the HTTP server
func (s Server) Serve(port int) {
	portString := ":" + strconv.Itoa(port)
	log.Println("Listening on http://localhost" + portString)
	err := http.ListenAndServe(portString, handlers.LoggingHandler(os.Stdout, s.router))
	if err != nil {
		log.Println("Error starting server: ", err)
	}
}

// Index says hello
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

const feedTemplate = "server/get_feed.gotmpl"

// GetFeed returns useful info about a feed
func (s Server) GetFeed(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	feedName := ps.ByName("name")

	if _, ok := s.Feeds[feedName]; !ok {
		http.Error(w, fmt.Sprintf("Not Found: Couldn't find feed %q", feedName), http.StatusUnauthorized)
		return
	}

	t, err := template.ParseFiles(feedTemplate)
	if err != nil {
		panic(err)
	}

	digests, err := s.db.GetDigestsByFeed(feedName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: Failed to get digests: %s", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Name    string
		Digests []models.Digest
	}{feedName, digests}

	err = t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

// GetFeedRSS returns the RSS feed
func (s Server) GetFeedRSS(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	feedName := ps.ByName("name")

	if _, ok := s.Feeds[feedName]; !ok {
		http.Error(w, fmt.Sprintf("Not Found: Couldn't find feed %q", feedName), http.StatusUnauthorized)
		return
	}

	digests, err := s.db.GetDigestsByFeed(feedName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: Failed to get digests: %s", err), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("Mailshine - %s", s.Feeds[feedName].Title),
		Link:        &feeds.Link{Href: "http://localhost:8080/feeds/" + feedName + "/rss"},
		Description: "cool feed with cooler content",
		// Author:      &feeds.Author{Name: "Jason Moiron", Email: "jmoiron@jmoiron.net"},
		Created: now,
	}

	for _, digest := range digests {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       digest.Title,
			Link:        &feeds.Link{Href: "http://jmoiron.net/blog/limiting-concurrency-in-go/"},
			Description: renderDigest(digest, s.baseURL),
			// Author:      &feeds.Author{Name: "Jason Moiron", Email: "jmoiron@jmoiron.net"},
			Created: digest.CreatedAt,
		})
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte(rss))
}

// GetPreviewRSS previews the HTML description for a feed
func (s Server) GetPreviewRSS(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	feedName := ps.ByName("name")

	if _, ok := s.Feeds[feedName]; !ok {
		http.Error(w, fmt.Sprintf("Not Found: Couldn't find feed %q", feedName), http.StatusUnauthorized)
		return
	}

	digests, err := s.db.GetDigestsByFeed(feedName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: Failed to get digests: %s", err), http.StatusInternalServerError)
		return
	}

	if len(digests) == 0 {
		http.Error(w, ("Error: No digests found to preview"), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(renderDigest(digests[0], s.baseURL)))
}

const feedRSSTemplate = "server/digest.gotmpl"

// renderDigest renders the HTML representation of a digest
// TODO: Load the templates once in initialization, instead of every render
func renderDigest(digest models.Digest, baseURL string) string {
	t, err := template.New(path.Base(feedRSSTemplate)).Funcs(
		template.FuncMap{
			"trimWww": func(s string) string {
				return strings.TrimPrefix(s, "www.")
			},
		}).ParseFiles(feedRSSTemplate)
	if err != nil {
		log.Printf("Failed to parse template: %s", err)
	}

	tmplData := struct {
		BaseURL string
		Digest  models.Digest
	}{baseURL, digest}

	var buff bytes.Buffer
	err = t.Execute(&buff, tmplData)
	if err != nil {
		log.Printf("Failed to render digest: %s", err)
	}

	return buff.String()
}
