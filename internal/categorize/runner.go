package categorize

import (
	"context"
	"sync"

	"github.com/sh-yazdipour/vibe-badget/internal/store"
)

type Classifier interface {
	Classify(ctx context.Context, partner string, categories []string) (category string, reason string, err error)
}

// LogEntry records why a single transaction ended up the way it did during a
// categorize run, for display in the frontend's "Run AI categorization" log.
type LogEntry struct {
	TxID     int64  `json:"tx_id"`
	Partner  string `json:"partner"`
	Category string `json:"category"`
	Source   string `json:"source"` // "rule", "llm", or "skipped"
	Reason   string `json:"reason"`
}

type Result struct {
	Rules   int        `json:"rules"`
	LLM     int        `json:"llm"`
	Skipped int        `json:"skipped"`
	Log     []LogEntry `json:"log"`
}

// Run applies rules to every uncategorized transaction, then sends whatever
// is left to the LLM through a bounded worker pool of size `concurrency`.
// onEntry, if non-nil, is called once per transaction the instant it's
// resolved (rule match, LLM result, or skip), in addition to being recorded
// in the returned Result.Log — used to stream progress to a caller.
func Run(ctx context.Context, s *store.Store, llm Classifier, concurrency int, onEntry func(LogEntry)) (Result, error) {
	res := Result{Log: []LogEntry{}}
	var mu sync.Mutex
	record := func(e LogEntry) {
		res.Log = append(res.Log, e)
		if onEntry != nil {
			onEntry(e)
		}
	}

	rules, err := s.ActiveRules()
	if err != nil {
		return res, err
	}
	byName, names, err := s.CategoryNames()
	if err != nil {
		return res, err
	}
	idToName := map[int64]string{}
	for name, id := range byName {
		idToName[id] = name
	}

	var classifiableNames []string
	for _, n := range names {
		if n != "Ignore" {
			classifiableNames = append(classifiableNames, n)
		}
	}

	txns, err := s.UncategorizedTransactions()
	if err != nil {
		return res, err
	}

	// Pass 1: rules (cheap, sequential — no locking needed here).
	var forLLM []int64
	partnerOf := map[int64]string{}
	for _, t := range txns {
		if catID, ok := Match(t, rules); ok {
			if err := s.SetCategory(t.ID, catID, "rule"); err != nil {
				return res, err
			}
			res.Rules++
			record(LogEntry{
				TxID: t.ID, Partner: t.PartnerName, Category: idToName[catID],
				Source: "rule", Reason: "matched rule",
			})
			continue
		}
		forLLM = append(forLLM, t.ID)
		partnerOf[t.ID] = t.PartnerName
	}

	if llm == nil || concurrency < 1 {
		res.Skipped += len(forLLM)
		for _, id := range forLLM {
			record(LogEntry{
				TxID: id, Partner: partnerOf[id], Source: "skipped", Reason: "no LLM configured",
			})
		}
		return res, nil
	}

	// Pass 2: LLM in parallel.
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for _, id := range forLLM {
		wg.Add(1)
		sem <- struct{}{}
		go func(txID int64) {
			defer wg.Done()
			defer func() { <-sem }()

			name, reason, cerr := llm.Classify(ctx, partnerOf[txID], classifiableNames)
			mu.Lock()
			defer mu.Unlock()
			if cerr != nil {
				res.Skipped++
				record(LogEntry{
					TxID: txID, Partner: partnerOf[txID], Source: "skipped", Reason: cerr.Error(),
				})
				return
			}
			if name == "Uncategorized" {
				res.Skipped++
				skipReason := reason
				if skipReason == "" {
					skipReason = "LLM returned Uncategorized"
				}
				record(LogEntry{
					TxID: txID, Partner: partnerOf[txID], Source: "skipped", Reason: skipReason,
				})
				return
			}
			if catID, ok := byName[name]; ok {
				if err := s.SetCategory(txID, catID, "llm"); err == nil {
					res.LLM++
					record(LogEntry{
						TxID: txID, Partner: partnerOf[txID], Category: name,
						Source: "llm", Reason: reason,
					})
					return
				}
			}
			res.Skipped++
			record(LogEntry{
				TxID: txID, Partner: partnerOf[txID], Source: "skipped", Reason: "unknown category returned: " + name,
			})
		}(id)
	}
	wg.Wait()
	return res, nil
}
