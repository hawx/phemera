package main

import (
	"flag"
	"fmt"
	clear "github.com/gorilla/context"
	"github.com/hawx/phemera/cookie"
	database "github.com/hawx/phemera/db"
	"github.com/hawx/phemera/markdown"
	"github.com/hawx/phemera/models"
	"github.com/hawx/phemera/persona"
	"github.com/hoisie/mustache"
	"github.com/stvp/go-toml-config"
	"log"
	"net/http"
	"time"
)

var store cookie.Store

var (
	assetDir     = flag.String("assets", "./assets", "")
	settingsPath = flag.String("settings", "./settings.toml", "")
	port         = flag.String("port", "8080", "")

	title        = config.String("title", "'phemera")
	description  = config.String("description", "this is is is")
	url          = config.String("url", "http://localhost:8080")
	user         = config.String("user", "someone@example.com")
	cookieSecret = config.String("secret", "some-secret")
	audience     = config.String("audience", "localhost")
	horizon      = config.String("horizon", "248400s")
	dbPath       = config.String("db", "./db")
)

type Context struct {
	Entries  models.Entries
	LoggedIn bool

	Title       string
	Description string
	Url         string
}

func ctx(db database.Db, r *http.Request) Context {
	return Context{
		Entries:     db.Get(),
		LoggedIn:    LoggedIn(r),
		Title:       *title,
		Description: markdown.Render(*description),
		Url:         *url,
	}
}

func LoggedIn(r *http.Request) bool {
	return store.Get(r) == *user
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()

	if err := config.Parse(*settingsPath); err != nil {
		log.Fatal("toml:", err)
	}

	store = cookie.NewStore(*cookieSecret)

	db := database.Open(*dbPath, *horizon)
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			body := mustache.RenderFileInLayout("views/list.mustache", "views/layout.mustache", ctx(db, r))

			w.Header().Add("Content-Type", "text/html")
			fmt.Fprint(w, body)
		}
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if !LoggedIn(r) {
				w.WriteHeader(403)
				return
			}

			body := mustache.RenderFileInLayout("views/add.mustache", "views/layout.mustache")
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprintf(w, body)
			return
		}

		if r.Method == "POST" {
			if !LoggedIn(r) {
				w.WriteHeader(403)
				return
			}

			db.Save(time.Now(), r.PostFormValue("body"))
			http.Redirect(w, r, "/", 301)
			return
		}
	})

	http.HandleFunc("/feed", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			body := mustache.RenderFile("views/feed.mustache", ctx(db, r))

			w.Header().Add("Content-Type", "application/rss+xml")
			fmt.Fprintf(w, body)
		}
	})

	http.HandleFunc("/sign-in", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			assertion := r.PostFormValue("assertion")
			email, err := persona.Assert(*audience, assertion)

			if err != nil {
				log.Print("sign-in:", err)
				w.WriteHeader(403)
				return
			}

			store.Set(email, w, r)
			w.WriteHeader(200)
		}
	})

	http.HandleFunc("/sign-out", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			store.Set("-", w, r)
			http.Redirect(w, r, "/", 307)
		}
	})

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(*assetDir))))

	log.Print("Running on :" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, clear.ClearHandler(Log(http.DefaultServeMux))))
}
