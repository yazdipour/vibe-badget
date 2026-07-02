package httpapi

import (
	"net/http"
	"strings"

	"github.com/sh-yazdipour/vibe-badget/internal/categorize"
)

type aiRuleSuggestionResponse struct {
	Pattern      string `json:"pattern"`
	MatchType    string `json:"match_type"`
	CategoryName string `json:"category_name"`
	CategoryID   int64  `json:"category_id"`
	Reason       string `json:"reason"`
}

func (s *Server) suggestRules(w http.ResponseWriter, r *http.Request) {
	kv, err := s.store.GetSettings()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if kv["llm_base_url"] == "" || kv["llm_model"] == "" {
		http.Error(w, "LLM not configured", 400)
		return
	}
	llm := categorize.NewLLM(categorize.LLMConfig{
		BaseURL: kv["llm_base_url"], APIKey: kv["llm_api_key"], Model: kv["llm_model"],
	})

	txns, err := s.store.ListTransactions(0)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	rules, err := s.store.ActiveRules()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	cats, err := s.store.ListCategories()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	covered := map[string]bool{}
	var existingPatterns []string
	for _, rule := range rules {
		existingPatterns = append(existingPatterns, rule.Pattern)
		if rule.Field == "partner_name" && rule.MatchType == "exact" {
			covered[strings.ToLower(strings.TrimSpace(rule.Pattern))] = true
		}
	}

	// Determine each uncovered partner's most-common LLM-assigned category.
	type tally struct {
		counts map[string]int
		order  []string
	}
	partnerCats := map[string]*tally{}
	var partnerOrder []string
	for _, t := range txns {
		if t.CategorizedBy != "llm" || t.CategoryName == "" {
			continue
		}
		tl, ok := partnerCats[t.PartnerName]
		if !ok {
			tl = &tally{counts: map[string]int{}}
			partnerCats[t.PartnerName] = tl
			partnerOrder = append(partnerOrder, t.PartnerName)
		}
		if tl.counts[t.CategoryName] == 0 {
			tl.order = append(tl.order, t.CategoryName)
		}
		tl.counts[t.CategoryName]++
	}

	var partners []categorize.PartnerCategory
	for _, partner := range partnerOrder {
		if covered[strings.ToLower(strings.TrimSpace(partner))] {
			continue
		}
		tl := partnerCats[partner]
		best := tl.order[0]
		for _, cat := range tl.order {
			if tl.counts[cat] > tl.counts[best] {
				best = cat
			}
		}
		partners = append(partners, categorize.PartnerCategory{Partner: partner, Category: best})
	}

	var categoryNames []string
	catByName := map[string]int64{}
	for _, c := range cats {
		categoryNames = append(categoryNames, c.Name)
		catByName[c.Name] = c.ID
	}

	suggestions, err := llm.SuggestRules(r.Context(), partners, existingPatterns, categoryNames)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out := []aiRuleSuggestionResponse{}
	for _, sug := range suggestions {
		pattern := strings.TrimSpace(sug.Pattern)
		if pattern == "" || covered[strings.ToLower(pattern)] {
			continue
		}
		catID, ok := catByName[sug.Category]
		if !ok {
			continue
		}
		out = append(out, aiRuleSuggestionResponse{
			Pattern: sug.Pattern, MatchType: sug.MatchType, CategoryName: sug.Category,
			CategoryID: catID, Reason: sug.Reason,
		})
	}
	writeJSON(w, 200, out)
}
