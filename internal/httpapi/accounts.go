package httpapi

import (
	"net/http"
	"strconv"
)

func (s *Server) listAccounts(w http.ResponseWriter, r *http.Request) {
	accs, err := s.store.ListAccounts()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 200, accs)
}

func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", 400)
		return
	}
	if err := s.store.DeleteAccount(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}
