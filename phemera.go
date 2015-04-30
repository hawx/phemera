package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/antage/eventsource"
	"github.com/hawx/persona"
	"github.com/hawx/serve"
	"github.com/stvp/go-toml-config"
	"hawx.me/code/mux"

	"hawx.me/code/phemera/assets"
	database "hawx.me/code/phemera/db"
	"hawx.me/code/phemera/markdown"
	"hawx.me/code/phemera/models"
	"hawx.me/code/phemera/views"
)

var (
	settingsPath = flag.String("settings", "./settings.toml", "")
	port         = flag.String("port", "8080", "")
	socket       = flag.String("socket", "", "")

	title        = config.String("title", "'phemera")
	description  = config.String("description", "['phemera](https://hawx.me/code/phemera) is an experiment in forgetful blogging.")
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

func ctx(db database.Db, r *http.Request, loggedIn bool) Context {
	return Context{
		Entries:     db.Get(),
		LoggedIn:    loggedIn,
		Title:       *title,
		Description: markdown.Render(*description),
		SafeDesc:    *description,
		Url:         *url,
	}
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func List(db database.Db, loggedIn bool) http.Handler {
	return mux.Method{
		"GET": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := views.List.RenderInLayout(views.Layout, ctx(db, r, loggedIn))
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprintf(w, body)
		}),
	}
}

func Add(db database.Db, es eventsource.EventSource) http.Handler {
	return mux.Method{
		"GET": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := views.Add.RenderInLayout(views.Layout, ctx(db, r, true))
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprintf(w, body)
		}),
		"POST": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			db.Save(time.Now(), r.PostFormValue("body"))
			es.SendEventMessage(r.PostFormValue("body"), "add-post", "")
			http.Redirect(w, r, "/", 301)
		}),
	}
}

var Preview = mux.Method{
	"POST": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		markdown.RenderTo(r.Body, w)
	}),
}

func Feed(db database.Db) http.Handler {
	return mux.Method{
		"GET": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := views.Feed.Render(ctx(db, r, false))

			w.Header().Add("Content-Type", "application/rss+xml")
			fmt.Fprintf(w, body)
		}),
	}
}

func main() {
	flag.Parse()

	if err := config.Parse(*settingsPath); err != nil {
		log.Fatal("toml:", err)
	}

	db := database.Open(*dbPath, *horizon)
	defer db.Close()

	es := eventsource.New(nil, nil)
	defer es.Close()

	store := persona.NewStore(*cookieSecret)
	persona := persona.New(store, *audience, []string{*user})

	http.Handle("/", persona.Switch(List(db, true), List(db, false)))
	http.Handle("/add", persona.Protect(Add(db, es)))
	http.Handle("/preview", persona.Protect(Preview))
	http.Handle("/feed", Feed(db))
	http.Handle("/connect", es)

	http.Handle("/sign-in", mux.Method{"POST": persona.SignIn})
	http.Handle("/sign-out", mux.Method{"GET": persona.SignOut})

	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Server(map[string]string{
		"jquery.caret.js":        assets.Caret,
		"jquery.autosize.min.js": assets.AutoSize,
		"styles.css":             assets.Styles,
		"list.js":                assets.List,
	})))

	serve.Serve(*port, *socket, Log(http.DefaultServeMux))
}
