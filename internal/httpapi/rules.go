package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sh-yazdipour/vibe-wallet/internal/model"
)

var validFields = map[string]bool{"partner_iban": true, "partner_name": true, "type": true, "payment_reference": true}
var validMatch = map[string]bool{"exact": true, "keyword": true}

func (s *Server) listRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.store.ListRules()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 200, rules)
}

func (s *Server) createRule(w http.ResponseWriter, r *http.Request) {
	var in model.Rule
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if !validFields[in.Field] || !validMatch[in.MatchType] || in.Pattern == "" || in.CategoryID == 0 {
		http.Error(w, "invalid rule", 400)
		return
	}
	out, err := s.store.CreateRule(in)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 201, out)
}

func (s *Server) deleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", 400)
		return
	}
	if err := s.store.DeleteRule(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}
