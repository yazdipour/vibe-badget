package categorize

import (
	"testing"

	"github.com/sh-yazdipour/vibe-badget/internal/model"
)

func TestMatchPriority(t *testing.T) {
	rules := []model.Rule{
		{Field: "payment_reference", MatchType: "keyword", Pattern: "round", CategoryID: 99},
		{Field: "partner_name", MatchType: "keyword", Pattern: "lidl", CategoryID: 1},
		{Field: "partner_iban", MatchType: "exact", Pattern: "AT999", CategoryID: 7},
	}

	// partner_iban wins over everything else
	got, ok := Match(model.Transaction{PartnerIban: "AT999", PartnerName: "LIDL", PaymentReference: "round-up"}, rules)
	if !ok || got != 7 {
		t.Fatalf("iban priority: got %d ok %v", got, ok)
	}

	// no iban -> partner_name keyword
	got, ok = Match(model.Transaction{PartnerName: "LIDL DANKT", PaymentReference: "round-up"}, rules)
	if !ok || got != 1 {
		t.Fatalf("name keyword: got %d ok %v", got, ok)
	}

	// nothing matches
	if _, ok := Match(model.Transaction{PartnerName: "Unknown"}, rules); ok {
		t.Fatal("want no match")
	}
}
