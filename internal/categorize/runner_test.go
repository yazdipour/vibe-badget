package categorize

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/sh-yazdipour/vibe-badget/internal/db"
	"github.com/sh-yazdipour/vibe-badget/internal/model"
	"github.com/sh-yazdipour/vibe-badget/internal/store"
)

type fakeLLM struct {
	mu    sync.Mutex
	calls [][]string
}

func (f *fakeLLM) ClassifyBatch(_ context.Context, partners []string, _ []string) (map[string]BatchClassification, error) {
	f.mu.Lock()
	f.calls = append(f.calls, append([]string(nil), partners...))
	f.mu.Unlock()
	out := map[string]BatchClassification{}
	for _, p := range partners {
		out[p] = BatchClassification{Category: "Transport"}
	}
	return out, nil
}

func (f *fakeLLM) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
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
	if res.Rules != 1 || res.LLM != 1 || f.callCount() != 1 {
		t.Fatalf("unexpected result %+v llmCalls=%d", res, f.callCount())
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
			if e.Partner != "Mystery Cab Co" || e.Category != "Transport" {
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

type recordingLLM struct {
	lastCategories []string
}

func (f *recordingLLM) ClassifyBatch(_ context.Context, partners []string, categories []string) (map[string]BatchClassification, error) {
	f.lastCategories = categories
	out := map[string]BatchClassification{}
	for _, p := range partners {
		out[p] = BatchClassification{Category: "Uncategorized"}
	}
	return out, nil
}

func TestRunNeverOffersIgnoreToLLM(t *testing.T) {
	d, _ := db.Open(":memory:")
	defer d.Close()
	s := store.New(d)

	_, err := s.InsertTransactions([]model.Transaction{
		{AccountName: "Main", PartnerName: "Mystery Shop", DedupeHash: "ignore-test-1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	f := &recordingLLM{}
	_, err = Run(context.Background(), s, f, 1, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	for _, c := range f.lastCategories {
		if c == "Ignore" {
			t.Fatalf("Ignore must not be offered to the LLM, got categories: %v", f.lastCategories)
		}
	}
}

func TestRunDedupesSamePartnerIntoOneLLMCall(t *testing.T) {
	d, _ := db.Open(":memory:")
	defer d.Close()
	s := store.New(d)

	_, err := s.InsertTransactions([]model.Transaction{
		{AccountName: "Main", PartnerName: "Mystery Cab Co", DedupeHash: "dedupe-1"},
		{AccountName: "Main", PartnerName: "Mystery Cab Co", DedupeHash: "dedupe-2"},
	})
	if err != nil {
		t.Fatal(err)
	}

	f := &fakeLLM{}
	res, err := Run(context.Background(), s, f, 4, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.LLM != 2 {
		t.Fatalf("want both transactions categorized, got LLM=%d", res.LLM)
	}
	if f.callCount() != 1 {
		t.Fatalf("want 1 deduped LLM call for 2 transactions sharing a partner, got %d calls: %+v", f.callCount(), f.calls)
	}
	if len(f.calls[0]) != 1 {
		t.Fatalf("want the single call to list exactly 1 unique partner, got %v", f.calls[0])
	}
}

func TestRunBatchesMoreThan50UniquePartners(t *testing.T) {
	d, _ := db.Open(":memory:")
	defer d.Close()
	s := store.New(d)

	var txns []model.Transaction
	for i := 0; i < 51; i++ {
		txns = append(txns, model.Transaction{
			AccountName: "Main",
			PartnerName: fmt.Sprintf("Unique Partner %d", i),
			DedupeHash:  fmt.Sprintf("batch-%d", i),
		})
	}
	if _, err := s.InsertTransactions(txns); err != nil {
		t.Fatal(err)
	}

	f := &fakeLLM{}
	res, err := Run(context.Background(), s, f, 1, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.LLM != 51 {
		t.Fatalf("want all 51 transactions categorized, got LLM=%d", res.LLM)
	}
	if f.callCount() != 2 {
		t.Fatalf("want 51 unique partners split into 2 batches of <=50, got %d calls", f.callCount())
	}
	total := 0
	for _, c := range f.calls {
		total += len(c)
	}
	if total != 51 {
		t.Fatalf("want 51 total partners across all batch calls, got %d", total)
	}
}
