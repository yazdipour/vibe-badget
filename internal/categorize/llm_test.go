package categorize

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClassifyBatchParsesResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `[
					{"partner":"LIDL DANKT","category":"Groceries"},
					{"partner":"Taxi Co","category":"Transport"}
				]`}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	results, err := llm.ClassifyBatch(context.Background(), []string{"LIDL DANKT", "Taxi Co"}, []string{"Groceries", "Transport"})
	if err != nil {
		t.Fatalf("ClassifyBatch: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d: %+v", len(results), results)
	}
	if r := results["LIDL DANKT"]; r.Category != "Groceries" {
		t.Fatalf("unexpected LIDL DANKT result: %+v", r)
	}
	if r := results["Taxi Co"]; r.Category != "Transport" {
		t.Fatalf("unexpected Taxi Co result: %+v", r)
	}
}

func TestClassifyBatchStripsMarkdownFences(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "```json\n[{\"partner\":\"Taxi Co\",\"category\":\"Transport\"}]\n```"}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	results, err := llm.ClassifyBatch(context.Background(), []string{"Taxi Co"}, []string{"Transport"})
	if err != nil || results["Taxi Co"].Category != "Transport" {
		t.Fatalf("got results=%+v err=%v", results, err)
	}
}

func TestClassifyBatchMissingPartnerOmitted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `[{"partner":"LIDL DANKT","category":"Groceries"}]`}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	results, err := llm.ClassifyBatch(context.Background(), []string{"LIDL DANKT", "Unanswered Partner"}, []string{"Groceries"})
	if err != nil {
		t.Fatalf("ClassifyBatch: %v", err)
	}
	if _, ok := results["Unanswered Partner"]; ok {
		t.Fatalf("expected no entry for a partner the LLM didn't answer, got: %+v", results)
	}
	if results["LIDL DANKT"].Category != "Groceries" {
		t.Fatalf("unexpected result: %+v", results)
	}
}

func TestClassifyBatchUnknownCategoryFallsBackToUncategorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `[{"partner":"???","category":"Spaceships"}]`}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	results, err := llm.ClassifyBatch(context.Background(), []string{"???"}, []string{"Groceries"})
	if err != nil {
		t.Fatalf("ClassifyBatch: %v", err)
	}
	if results["???"].Category != "Uncategorized" {
		t.Fatalf("want Uncategorized, got %+v", results["???"])
	}
}

func TestPingOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"id": "gemma3:12b"}, {"id": "llama3.1"}},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "gemma3:12b"})
	res := llm.Ping(context.Background())
	if res.Status != "ok" {
		t.Fatalf("want ok, got %+v", res)
	}
}

func TestPingModelNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"id": "llama3.1"}},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "gemma3:12b"})
	res := llm.Ping(context.Background())
	if res.Status != "model_not_found" {
		t.Fatalf("want model_not_found, got %+v", res)
	}
}

func TestPingUnreachable(t *testing.T) {
	llm := NewLLM(LLMConfig{BaseURL: "http://127.0.0.1:1", Model: "gemma3:12b"})
	res := llm.Ping(context.Background())
	if res.Status != "unreachable" {
		t.Fatalf("want unreachable, got %+v", res)
	}
}

func TestSuggestRulesParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `[{"pattern":"AMAZON","match_type":"keyword","category":"Shopping","reason":"varies per order"}]`}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	suggestions, err := llm.SuggestRules(context.Background(),
		[]PartnerCategory{{Partner: "AMAZON MKTP DE", Category: "Shopping"}},
		nil, []string{"Shopping", "Groceries"})
	if err != nil {
		t.Fatalf("SuggestRules: %v", err)
	}
	if len(suggestions) != 1 || suggestions[0].Pattern != "AMAZON" || suggestions[0].MatchType != "keyword" ||
		suggestions[0].Category != "Shopping" || suggestions[0].Reason != "varies per order" {
		t.Fatalf("unexpected suggestions: %+v", suggestions)
	}
}

func TestSuggestRulesDefaultsInvalidMatchType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `[{"pattern":"Lidl","match_type":"fuzzy","category":"Groceries","reason":"x"}]`}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	suggestions, err := llm.SuggestRules(context.Background(),
		[]PartnerCategory{{Partner: "Lidl", Category: "Groceries"}}, nil, []string{"Groceries"})
	if err != nil {
		t.Fatalf("SuggestRules: %v", err)
	}
	if len(suggestions) != 1 || suggestions[0].MatchType != "keyword" {
		t.Fatalf("expected match_type to default to keyword: %+v", suggestions)
	}
}

func TestSuggestRulesDropsUnknownCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `[
					{"pattern":"Lidl","match_type":"exact","category":"Groceries","reason":"x"},
					{"pattern":"Ghost","match_type":"exact","category":"NotARealCategory","reason":"y"}
				]`}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	suggestions, err := llm.SuggestRules(context.Background(),
		[]PartnerCategory{{Partner: "Lidl", Category: "Groceries"}, {Partner: "Ghost", Category: "NotARealCategory"}},
		nil, []string{"Groceries"})
	if err != nil {
		t.Fatalf("SuggestRules: %v", err)
	}
	if len(suggestions) != 1 || suggestions[0].Pattern != "Lidl" {
		t.Fatalf("expected only the known-category suggestion to survive: %+v", suggestions)
	}
}
