package categorize

import (
	"strings"

	"github.com/sh-yazdipour/vibe-badget/internal/model"
)

var fieldPriority = []string{"partner_iban", "partner_name", "type", "payment_reference"}

func fieldValue(t model.Transaction, field string) string {
	switch field {
	case "partner_iban":
		return t.PartnerIban
	case "partner_name":
		return t.PartnerName
	case "type":
		return t.Type
	case "payment_reference":
		return t.PaymentReference
	}
	return ""
}

// Match returns the category id of the first rule that matches, honouring
// field priority (iban > name > type > reference) and exact-before-keyword.
func Match(t model.Transaction, rules []model.Rule) (int64, bool) {
	for _, field := range fieldPriority {
		val := strings.ToLower(strings.TrimSpace(fieldValue(t, field)))
		if val == "" {
			continue
		}
		// exact first
		for _, r := range rules {
			if r.Field == field && r.MatchType == "exact" &&
				val == strings.ToLower(strings.TrimSpace(r.Pattern)) {
				return r.CategoryID, true
			}
		}
		// then keyword
		for _, r := range rules {
			if r.Field == field && r.MatchType == "keyword" &&
				strings.Contains(val, strings.ToLower(strings.TrimSpace(r.Pattern))) {
				return r.CategoryID, true
			}
		}
	}
	return 0, false
}
