package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/hawx/phemera/assets"
	"github.com/hawx/phemera/cookie"
	database "github.com/hawx/phemera/db"
	"github.com/hawx/phemera/markdown"
	"github.com/hawx/phemera/models"
	"github.com/hawx/phemera/persona"
	"github.com/hawx/phemera/views"
	"github.com/hoisie/mustache"
	"github.com/stvp/go-toml-config"
	"log"
	"net/http"
	"time"
)

var store cookie.Store

var (
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
	SafeDesc    string
	Url         string
}

func ctx(db database.Db, r *http.Request) Context {
	return Context{
		Entries:     db.Get(),
		LoggedIn:    LoggedIn(r),
		Title:       *title,
		Description: markdown.Render(*description),
	  SafeDesc:    *description,
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

func Render(template *mustache.Template, db database.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := template.RenderInLayout(views.Layout, ctx(db, r))
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

	protect := persona.Protector(store, []string{*user})

	r.Path("/").Methods("GET").Handler(Render(views.List, db))
	r.Path("/add").Methods("GET").Handler(protect(Render(views.Add, db)))
	r.Path("/add").Methods("POST").Handler(protect(Add(db)))

	r.Path("/feed").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := views.Feed.Render(ctx(db, r))

		w.Header().Add("Content-Type", "application/rss+xml")
		fmt.Fprintf(w, body)
	})

	r.Path("/sign-in").Methods("POST").Handler(persona.SignIn(store, *audience))
	r.Path("/sign-out").Methods("GET").Handler(persona.SignOut(store))

	http.Handle("/", r)
	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Server(map[string]string{
		"jquery.caret.js":        assets.Caret,
		"jquery.autosize.min.js": assets.AutoSize,
		"styles.css":             assets.Styles,
		"list.js":                assets.List,
	})))

	log.Print("Running on :" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, context.ClearHandler(Log(http.DefaultServeMux))))
}
