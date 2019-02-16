package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/antage/eventsource"
	"github.com/gorilla/sessions"
	"hawx.me/code/indieauth"
	"hawx.me/code/mux"
	"hawx.me/code/serve"

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
	Me          string
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

	if conf.Me == "" {
		log.Fatal("me must be set to sign-in")
	}

	auth, err := indieauth.Authentication(conf.URL, conf.URL+"/callback")
	if err != nil {
		log.Fatal(err)
	}

	endpoints, err := indieauth.FindEndpoints(conf.Me)
	if err != nil {
		log.Fatal(err)
	}

	db := database.Open(conf.DbPath, conf.Horizon)
	defer db.Close()

	es := eventsource.New(nil, nil)
	defer es.Close()

	store := NewStore(conf.Secret)

	protect := func(store *meStore, good, bad http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if addr := store.Get(r); addr == conf.Me {
				good.ServeHTTP(w, r)
			} else {
				bad.ServeHTTP(w, r)
			}
		})
	}

	http.Handle("/", protect(store, List(conf, db, true), List(conf, db, false)))
	http.Handle("/add", protect(store, Add(conf, db, es), http.NotFoundHandler()))
	http.Handle("/preview", protect(store, Preview, http.NotFoundHandler()))
	http.Handle("/feed", Feed(conf, db))
	http.Handle("/connect", es)

	http.HandleFunc("/sign-in", func(w http.ResponseWriter, r *http.Request) {
		state, err := store.SetState(w, r)
		if err != nil {
			http.Error(w, "could not start auth", http.StatusInternalServerError)
			return
		}

		redirectURL := auth.RedirectURL(endpoints, conf.Me, state)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		state := store.GetState(r)

		if r.FormValue("state") != state {
			http.Error(w, "state is bad", http.StatusBadRequest)
			return
		}

		me, err := auth.Exchange(endpoints, r.FormValue("code"))
		if err != nil || me != conf.Me {
			http.Error(w, "nope", http.StatusForbidden)
			return
		}

		store.Set(w, r, me)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.HandleFunc("/sign-out", func(w http.ResponseWriter, r *http.Request) {
		store.Set(w, r, "")
		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.Handle("/assets/", http.StripPrefix("/assets/", assets.Server(map[string]string{
		"jquery.caret.js":        assets.Caret,
		"jquery.autosize.min.js": assets.AutoSize,
		"styles.css":             assets.Styles,
		"list.js":                assets.List,
	})))

	serve.Serve(*port, *socket, Log(http.DefaultServeMux))
}

type meStore struct {
	store sessions.Store
}

func NewStore(secret string) *meStore {
	return &meStore{sessions.NewCookieStore([]byte(secret))}
}

func (s meStore) Get(r *http.Request) string {
	session, _ := s.store.Get(r, "session")

	if v, ok := session.Values["me"].(string); ok {
		return v
	}

	return ""
}

func (s meStore) Set(w http.ResponseWriter, r *http.Request, me string) {
	session, _ := s.store.Get(r, "session")
	session.Values["me"] = me
	session.Save(r, w)
}

func (s meStore) SetState(w http.ResponseWriter, r *http.Request) (string, error) {
	bytes := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(bytes)

	session, _ := s.store.Get(r, "session")
	session.Values["state"] = state
	return state, session.Save(r, w)
}

func (s meStore) GetState(r *http.Request) string {
	session, _ := s.store.Get(r, "session")

	if v, ok := session.Values["state"].(string); ok {
		return v
	}

	return ""
}
