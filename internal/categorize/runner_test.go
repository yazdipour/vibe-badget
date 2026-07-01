package categorize

import (
	"context"
	"sync"
	"testing"

	"github.com/sh-yazdipour/vibe-badget/internal/db"
	"github.com/sh-yazdipour/vibe-badget/internal/model"
	"github.com/sh-yazdipour/vibe-badget/internal/store"
)

type fakeLLM struct{ called int }

func (f *fakeLLM) Classify(_ context.Context, partner string, _ []string) (string, string, error) {
	f.called++
	return "Transport", "looks like a taxi service", nil
}

func TestRunRulesThenLLM(t *testing.T) {
	d, _ := db.Open(":memory:")
	defer d.Close()
	s := store.New(d)

	// Lidl seed rule -> Groceries handles row 1; row 2 has no rule -> LLM.
	_, err := s.InsertTransactions([]model.Transaction{
		{AccountName: "Main", PartnerName: "LIDL DANKT", DedupeHash: "a"},
		{AccountName: "Main", PartnerName: "Mystery Cab Co", DedupeHash: "b"},
	})
	if err != nil {
		t.Fatal(err)
	}

	f := &fakeLLM{}
	var mu sync.Mutex
	var streamed []LogEntry
	onEntry := func(e LogEntry) {
		mu.Lock()
		defer mu.Unlock()
		streamed = append(streamed, e)
	}

	res, err := Run(context.Background(), s, f, 4, onEntry)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Rules != 1 || res.LLM != 1 || f.called != 1 {
		t.Fatalf("unexpected result %+v llmCalls=%d", res, f.called)
	}

	var remaining int
	d.QueryRow(`SELECT count(*) FROM transactions WHERE category_id IS NULL`).Scan(&remaining)
	if remaining != 0 {
		t.Fatalf("want 0 uncategorized, got %d", remaining)
	}

	if len(res.Log) != 2 {
		t.Fatalf("want 2 log entries, got %d: %+v", len(res.Log), res.Log)
	}
	if len(streamed) != len(res.Log) {
		t.Fatalf("onEntry callback count %d != res.Log count %d", len(streamed), len(res.Log))
	}
	var sawRule, sawLLM bool
	for _, e := range res.Log {
		switch e.Source {
		case "rule":
			sawRule = true
			if e.Partner != "LIDL DANKT" || e.Category != "Groceries" || e.Reason == "" {
				t.Fatalf("bad rule log entry: %+v", e)
			}
		case "llm":
			sawLLM = true
			if e.Partner != "Mystery Cab Co" || e.Category != "Transport" || e.Reason != "looks like a taxi service" {
				t.Fatalf("bad llm log entry: %+v", e)
			}
		default:
			t.Fatalf("unexpected log source: %+v", e)
		}
	}
	if !sawRule || !sawLLM {
		t.Fatalf("missing expected log sources: %+v", res.Log)
	}
}

func TestRunSkippedWithoutLLM(t *testing.T) {
	d, _ := db.Open(":memory:")
	defer d.Close()
	s := store.New(d)

	_, err := s.InsertTransactions([]model.Transaction{
		{AccountName: "Main", PartnerName: "Mystery Cab Co", DedupeHash: "b"},
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := Run(context.Background(), s, nil, 4, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Skipped != 1 || len(res.Log) != 1 || res.Log[0].Source != "skipped" {
		t.Fatalf("unexpected result %+v", res)
	}
}
