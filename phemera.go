package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/antage/eventsource"
	"hawx.me/code/mux"
	"hawx.me/code/serve"
	"hawx.me/code/uberich"

	"hawx.me/code/phemera/assets"
	database "hawx.me/code/phemera/db"
	"hawx.me/code/phemera/markdown"
	"hawx.me/code/phemera/models"
	"hawx.me/code/phemera/views"
)

type Conf struct {
	Title       string
	Description string
	URL         string
	Secret      string
	Horizon     string
	DbPath      string

	Uberich struct {
		AppName    string
		AppURL     string
		UberichURL string
		Secret     string
	}
}

type Context struct {
	Entries  models.Entries
	LoggedIn bool

	Title       string
	Description string
	SafeDesc    string
	Url         string
}

func ctx(conf *Conf, db database.Db, r *http.Request, loggedIn bool) Context {
	return Context{
		Entries:     db.Get(),
		LoggedIn:    loggedIn,
		Title:       conf.Title,
		Description: markdown.Render(conf.Description),
		SafeDesc:    conf.Description,
		Url:         conf.URL,
	}
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func List(conf *Conf, db database.Db, loggedIn bool) http.Handler {
	return mux.Method{
		"GET": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := views.List.RenderInLayout(views.Layout, ctx(conf, db, r, loggedIn))
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprintf(w, body)
		}),
	}
}

func Add(conf *Conf, db database.Db, es eventsource.EventSource) http.Handler {
	return mux.Method{
		"GET": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := views.Add.RenderInLayout(views.Layout, ctx(conf, db, r, true))
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

func Feed(conf *Conf, db database.Db) http.Handler {
	return mux.Method{
		"GET": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := views.Feed.Render(ctx(conf, db, r, false))

			w.Header().Add("Content-Type", "application/rss+xml")
			fmt.Fprintf(w, body)
		}),
	}
}

func main() {
	var (
		settingsPath = flag.String("settings", "./settings.toml", "")
		port         = flag.String("port", "8080", "")
		socket       = flag.String("socket", "", "")
	)
	flag.Parse()

	var conf *Conf
	if _, err := toml.DecodeFile(*settingsPath, &conf); err != nil {
		log.Fatal("toml:", err)
	}

	db := database.Open(conf.DbPath, conf.Horizon)
	defer db.Close()

	es := eventsource.New(nil, nil)
	defer es.Close()

	store := uberich.NewStore(conf.Secret)
	uberich := uberich.NewClient(conf.Uberich.AppName, conf.Uberich.AppURL, conf.Uberich.UberichURL, conf.Uberich.Secret, store)

	shield := func(h http.Handler) http.Handler {
		return uberich.Protect(h, http.NotFoundHandler())
	}

	http.Handle("/", uberich.Protect(List(conf, db, true), List(conf, db, false)))
	http.Handle("/add", shield(Add(conf, db, es)))
	http.Handle("/preview", shield(Preview))
	http.Handle("/feed", Feed(conf, db))
	http.Handle("/connect", es)

	http.Handle("/sign-in", uberich.SignIn("/"))
	http.Handle("/sign-out", uberich.SignOut("/"))

	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Server(map[string]string{
		"jquery.caret.js":        assets.Caret,
		"jquery.autosize.min.js": assets.AutoSize,
		"styles.css":             assets.Styles,
		"list.js":                assets.List,
	})))

	serve.Serve(*port, *socket, Log(http.DefaultServeMux))
}
