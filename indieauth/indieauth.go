package indieauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
)

type Store interface {
	Set(w http.ResponseWriter, r *http.Request, email string)
	Get(r *http.Request) string
}

type emailStore struct {
	store sessions.Store
}

func NewStore(secret string) Store {
	return &emailStore{sessions.NewCookieStore([]byte(secret))}
}

func (s emailStore) Get(r *http.Request) string {
	session, _ := s.store.Get(r, "session")

	if v, ok := session.Values["email"].(string); ok {
		return v
	}

	return ""
}

func (s emailStore) Set(w http.ResponseWriter, r *http.Request, email string) {
	session, _ := s.store.Get(r, "session")
	session.Values["email"] = email
	session.Save(r, w)
}

func SignIn() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		fmt.Fprint(w, `
<form action="https://indielogin.com/auth" method="get">
  <label for="url">Web Address:</label>
  <input id="url" type="text" name="me" placeholder="yourdomain.com" />
  <p><button type="submit">Sign In</button></p>
  <input type="hidden" name="client_id" value="http://localhost:8080/" />
  <input type="hidden" name="redirect_uri" value="http://localhost:8080/callback" />
  <input type="hidden" name="state" value="jwiusuerujs" />
</form>
`)
	})
}

func Callback(store Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("callback")

		state := r.FormValue("state")

		if state != "jwiusuerujs" {
			fmt.Fprint(w, "I don't know that state")
			return
		}

		code := r.FormValue("code")

		fmt.Println("making req")
		resp, err := http.PostForm("https://indielogin.com/auth", url.Values{
			"code":         {code},
			"redirect_uri": {"http://localhost:8080/callback"},
			"client_id":    {"http://localhost:8080/"},
		})

		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			io.Copy(w, resp.Body)
			return
		}

		var v struct {
			Me string `json:"me"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			fmt.Println("bad json")
			fmt.Fprint(w, err.Error())
			return
		}

		store.Set(w, r, v.Me)
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

func SignOut(store Store, redirectURI string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store.Set(w, r, "")
		http.Redirect(w, r, redirectURI, http.StatusFound)
	})
}

func Protect(store Store, good, bad http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if addr := store.Get(r); addr == "https://hawx.me/" {
			good.ServeHTTP(w, r)
		} else {
			bad.ServeHTTP(w, r)
		}
	})
}
