package categorize

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClassifyParsesCategoryAndReason(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `{"category":"Groceries","reason":"LIDL is a supermarket chain"}`}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	cat, reason, err := llm.Classify(context.Background(), "LIDL DANKT", []string{"Groceries", "Transport"})
	if err != nil || cat != "Groceries" || reason != "LIDL is a supermarket chain" {
		t.Fatalf("got cat=%q reason=%q err=%v", cat, reason, err)
	}
}

func TestClassifyStripsMarkdownFences(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "```json\n{\"category\":\"Transport\",\"reason\":\"taxi fare\"}\n```"}},
			},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	cat, reason, err := llm.Classify(context.Background(), "Taxi Co", []string{"Groceries", "Transport"})
	if err != nil || cat != "Transport" || reason != "taxi fare" {
		t.Fatalf("got cat=%q reason=%q err=%v", cat, reason, err)
	}
}

func TestClassifyFallsBackOnNonJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": "  Groceries\n"}}},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	cat, reason, err := llm.Classify(context.Background(), "LIDL DANKT", []string{"Groceries", "Transport"})
	if err != nil || cat != "Groceries" || reason != "" {
		t.Fatalf("got cat=%q reason=%q err=%v", cat, reason, err)
	}
}

func TestClassifyUnknownFallsBack(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": `{"category":"Spaceships","reason":"why not"}`}}},
		})
	}))
	defer srv.Close()

	llm := NewLLM(LLMConfig{BaseURL: srv.URL, Model: "test"})
	cat, _, err := llm.Classify(context.Background(), "???", []string{"Groceries"})
	if err != nil || cat != "Uncategorized" {
		t.Fatalf("want Uncategorized, got %q err %v", cat, err)
	}
}
