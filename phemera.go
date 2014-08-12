package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
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
	description  = config.String("description", "['phemera](https://github.com/hawx/phemera) is an experiment in forgetful blogging.")
	url          = config.String("url", "http://localhost:8080")
	user         = config.String("user", "someone@example.com")
	cookieSecret = config.String("secret", "some-secret-pls-change")
	audience     = config.String("audience", "localhost")
	horizon      = config.String("horizon", "-248400s")
	dbPath       = config.String("db", "./phemera-db")
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

func Protect(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !LoggedIn(r) {
			w.WriteHeader(403)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func Render(templatePath string, db database.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := mustache.RenderFileInLayout(templatePath, "views/layout.mustache", ctx(db, r))
		w.Header().Add("Content-Type", "text/html")
		fmt.Fprintf(w, body)
	})
}

func Add(db database.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db.Save(time.Now(), r.PostFormValue("body"))
		http.Redirect(w, r, "/", 301)
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

	r := mux.NewRouter()

	r.Path("/").Methods("GET").Handler(Render("views/list.mustache", db))
	r.Path("/add").Methods("GET").Handler(Protect(Render("views/add.mustache", db)))
	r.Path("/add").Methods("POST").Handler(Protect(Add(db)))

	r.Path("/feed").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := mustache.RenderFile("views/feed.mustache", ctx(db, r))

		w.Header().Add("Content-Type", "application/rss+xml")
		fmt.Fprintf(w, body)
	})

	r.Path("/sign-in").Methods("POST").Handler(persona.SignIn(store, *audience))
	r.Path("/sign-out").Methods("GET").Handler(persona.SignOut(store))

	http.Handle("/", r)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(*assetDir))))

	log.Print("Running on :" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, context.ClearHandler(Log(http.DefaultServeMux))))
}
