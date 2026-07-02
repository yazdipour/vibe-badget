package categorize

import (
	"context"
	"sync"

	"github.com/sh-yazdipour/vibe-badget/internal/store"
)

const classifyBatchSize = 50

type Classifier interface {
	ClassifyBatch(ctx context.Context, partners []string, categories []string) (map[string]BatchClassification, error)
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
// is left to the LLM. Transactions are deduplicated by partner name before
// hitting the LLM (identical partners share one classification), and unique
// partners are split into batches of classifyBatchSize sent as a single
// prompt each, fanned out through a bounded worker pool of size
// `concurrency`. onEntry, if non-nil, is called once per transaction the
// instant it's resolved (rule match, LLM result, or skip), in addition to
// being recorded in the returned Result.Log — used to stream progress to a
// caller. This per-transaction granularity is preserved even though the
// underlying LLM calls are now batched, so callers don't need to change.
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

	// Pass 2: dedupe by partner name, then classify in batches of
	// classifyBatchSize, fanned out through a bounded worker pool.
	partnerTxIDs := map[string][]int64{}
	var uniquePartners []string
	for _, id := range forLLM {
		p := partnerOf[id]
		if _, ok := partnerTxIDs[p]; !ok {
			uniquePartners = append(uniquePartners, p)
		}
		partnerTxIDs[p] = append(partnerTxIDs[p], id)
	}

	var batches [][]string
	for i := 0; i < len(uniquePartners); i += classifyBatchSize {
		end := i + classifyBatchSize
		if end > len(uniquePartners) {
			end = len(uniquePartners)
		}
		batches = append(batches, uniquePartners[i:end])
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for _, batch := range batches {
		wg.Add(1)
		sem <- struct{}{}
		go func(partners []string) {
			defer wg.Done()
			defer func() { <-sem }()

			results, berr := llm.ClassifyBatch(ctx, partners, classifiableNames)

			mu.Lock()
			defer mu.Unlock()
			for _, partner := range partners {
				ids := partnerTxIDs[partner]
				if berr != nil {
					for _, txID := range ids {
						res.Skipped++
						record(LogEntry{TxID: txID, Partner: partner, Source: "skipped", Reason: berr.Error()})
					}
					continue
				}
				result, ok := results[partner]
				if !ok {
					for _, txID := range ids {
						res.Skipped++
						record(LogEntry{TxID: txID, Partner: partner, Source: "skipped", Reason: "no result from LLM"})
					}
					continue
				}
				if result.Category == "Uncategorized" {
					for _, txID := range ids {
						res.Skipped++
						record(LogEntry{TxID: txID, Partner: partner, Source: "skipped", Reason: "LLM returned Uncategorized"})
					}
					continue
				}
				catID, ok := byName[result.Category]
				if !ok {
					for _, txID := range ids {
						res.Skipped++
						record(LogEntry{TxID: txID, Partner: partner, Source: "skipped", Reason: "unknown category returned: " + result.Category})
					}
					continue
				}
				for _, txID := range ids {
					if err := s.SetCategory(txID, catID, "llm"); err != nil {
						res.Skipped++
						record(LogEntry{TxID: txID, Partner: partner, Source: "skipped", Reason: err.Error()})
						continue
					}
					res.LLM++
					record(LogEntry{
						TxID: txID, Partner: partner, Category: result.Category,
						Source: "llm",
					})
				}
			}
		}(batch)
	}
	wg.Wait()
	return res, nil
}
