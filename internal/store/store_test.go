package store

import (
	"testing"

	"github.com/sh-yazdipour/vibe-badget/internal/db"
	"github.com/sh-yazdipour/vibe-badget/internal/model"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return New(d)
}

func TestInsertTransactionsIsIdempotent(t *testing.T) {
	s := newStore(t)
	txns := []model.Transaction{
		{AccountName: "Main", PartnerName: "LIDL", AmountEUR: -5, DedupeHash: "a"},
		{AccountName: "Main", PartnerName: "ALDI", AmountEUR: -9, DedupeHash: "b"},
	}
	n, err := s.InsertTransactions(txns)
	if err != nil || n != 2 {
		t.Fatalf("first insert: n=%d err=%v", n, err)
	}
	n, err = s.InsertTransactions(txns) // same rows again
	if err != nil || n != 0 {
		t.Fatalf("re-insert should be 0: n=%d err=%v", n, err)
	}
}
