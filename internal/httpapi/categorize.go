package httpapi

import (
	"context"
	"net/http"

	"github.com/sh-yazdipour/vibe-badget/internal/categorize"
)

func (s *Server) categorize(w http.ResponseWriter, r *http.Request) {
	res, err := s.runCategorize(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 200, res)
}

func (s *Server) runCategorize(ctx context.Context) (categorize.Result, error) {
	return categorize.Run(ctx, s.store, s.llm, 4)
}
