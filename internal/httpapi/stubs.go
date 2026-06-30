// internal/httpapi/stubs.go — temporary; replaced in Task 8.
package httpapi

import "net/http"

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) { http.Error(w, "todo", 501) }
func (s *Server) putSettings(w http.ResponseWriter, r *http.Request) { http.Error(w, "todo", 501) }
