package server

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/gorilla/handlers"
	"github.com/hebo/mailshine/models"
	"github.com/julienschmidt/httprouter"
)

// Server handles RSS and other HTTP routes
type Server struct {
	Feeds   models.FeedConfigMap
	router  *httprouter.Router
	db      models.DB
	baseURL string
}

// New creates a new Server
func New(db models.DB, feeds models.FeedConfigMap, baseURL string) Server {
	srv := Server{
		Feeds:   feeds,
		db:      db,
		baseURL: baseURL,
	}

	router := httprouter.New()

	router.GET("/", Index)
	router.GET("/feeds/:name", srv.GetFeed)
	router.GET("/feeds/:name/", srv.GetFeed)
	router.GET("/feeds/:name/rss", srv.GetFeedRSS)
	router.GET("/feeds/:name/digests/:digest_id", srv.GetDigest)

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

const templateGetFeed = "server/get_feed.gotmpl"

func digestURL(baseURL, feedName string, digestID int) string {
	return fmt.Sprintf("%s/feeds/%s/digests/%d", baseURL, feedName, digestID)
}

// GetFeed returns useful info about a feed
func (s Server) GetFeed(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	feedName := ps.ByName("name")

	if _, ok := s.Feeds[feedName]; !ok {
		http.Error(w, fmt.Sprintf("Not Found: Couldn't find feed %q", feedName), http.StatusUnauthorized)
		return
	}

	t, err := template.New(path.Base(templateGetFeed)).Funcs(
		template.FuncMap{
			"digestURL": func(feedName string, digestID int) string {
				return digestURL(s.baseURL, feedName, digestID)
			},
		}).ParseFiles(templateGetFeed)
	if err != nil {
		log.Printf("Failed to parse template: %s", err)
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
		log.Printf("Failed to render: %s", err)
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
		Created:     now,
	}

	for _, digest := range digests {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       digest.Title,
			Link:        &feeds.Link{Href: digestURL(s.baseURL, digest.FeedName, digest.ID)},
			Description: renderDigest(digest, s.baseURL),
			Created:     digest.CreatedAt,
		})
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Printf("Failed to render: %s", err)
	}

	w.Write([]byte(rss))
}

// GetDigest shows a single digest
func (s Server) GetDigest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	feedName := ps.ByName("name")
	digestID := ps.ByName("digest_id")

	if _, ok := s.Feeds[feedName]; !ok {
		http.Error(w, fmt.Sprintf("Not Found: Couldn't find feed %q", feedName), http.StatusUnauthorized)
		return
	}

	digest, err := s.db.GetDigestByID(digestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: Failed to get digest: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(renderDigest(digest, s.baseURL)))
}

const templateDigest = "server/digest.gotmpl"

// renderDigest renders the HTML representation of a digest
// TODO: Load the templates once in initialization, instead of every render
func renderDigest(digest models.Digest, baseURL string) string {
	t, err := template.New(path.Base(templateDigest)).Funcs(
		template.FuncMap{
			"trimWww": func(s string) string {
				return strings.TrimPrefix(s, "www.")
			},
			"trunc":      models.Truncate,
			"apolloLink": apolloURLHelper,
		}).ParseFiles(templateDigest)
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

func apolloURLHelper(s string) template.URL {
	u, _ := url.Parse(s)
	u.Scheme = "apollo"
	return template.URL(u.String())
}
