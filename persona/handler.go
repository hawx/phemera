package persona

import (
	"net/http"
	"log"
)

func SignIn(store emailStore, audience string) signInHandler {
	return signInHandler{store, audience}
}

func SignOut(store emailStore) signOutHandler {
	return signOutHandler{store}
}

type emailStore interface {
	Set(email string, w http.ResponseWriter, r *http.Request)
}

type signInHandler struct {
	store    emailStore
	audience string
}

func (s signInHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assertion := r.PostFormValue("assertion")
	email, err := Assert(s.audience, assertion)

	if err != nil {
		log.Print("persona:", err)
		w.WriteHeader(403)
		return
	}

	s.store.Set(email, w, r)
	w.WriteHeader(200)
}

type signOutHandler struct {
	store emailStore
}

func (s signOutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.store.Set("-", w, r)
	http.Redirect(w, r, "/", 307)
}
