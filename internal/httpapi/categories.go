package httpapi

import (
	"encoding/json"
	"net/http"
)

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := s.store.ListCategories()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 200, cats)
}

func (s *Server) createCategory(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name  string `json:"name"`
		Icon  string `json:"icon"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		http.Error(w, "name required", 400)
		return
	}
	icon := in.Icon
	if icon == "" {
		icon = "Tag"
	}
	color := in.Color
	if color == "" {
		color = "#6b7280"
	}
	c, err := s.store.CreateCategory(in.Name, icon, color)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 201, c)
}
