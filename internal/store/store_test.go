package store

import (
	"testing"

	"github.com/sh-yazdipour/vibe-wallet/internal/db"
	"github.com/sh-yazdipour/vibe-wallet/internal/model"
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

func TestCreateAndListCategoriesWithIconColor(t *testing.T) {
	s := newStore(t)

	c, err := s.CreateCategory("Pets", "PiggyBank", "#f59e0b", "#000000")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if c.Icon != "PiggyBank" || c.Color != "#f59e0b" || c.IconColor != "#000000" {
		t.Fatalf("unexpected created category: %+v", c)
	}

	cats, err := s.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	var found bool
	for _, cat := range cats {
		if cat.Name == "Pets" {
			found = true
			if cat.Icon != "PiggyBank" || cat.Color != "#f59e0b" || cat.IconColor != "#000000" {
				t.Fatalf("listed category mismatch: %+v", cat)
			}
		}
	}
	if !found {
		t.Fatal("Pets category not found in list")
	}

	// Re-creating with the same name upserts icon/color/icon_color instead of erroring.
	c2, err := s.CreateCategory("Pets", "Wallet", "#0ea5e9", "#ffffff")
	if err != nil {
		t.Fatalf("upsert CreateCategory: %v", err)
	}
	if c2.ID != c.ID || c2.Icon != "Wallet" || c2.Color != "#0ea5e9" || c2.IconColor != "#ffffff" {
		t.Fatalf("upsert did not update icon/color: %+v", c2)
	}
}

func TestUpdateCategoryAppearance(t *testing.T) {
	s := newStore(t)

	c, err := s.CreateCategory("Utilities2", "Tag", "#6b7280", "#ffffff")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	updated, err := s.UpdateCategoryAppearance(c.ID, "Zap", "#f59e0b", "#000000")
	if err != nil {
		t.Fatalf("UpdateCategoryAppearance: %v", err)
	}
	if updated.ID != c.ID || updated.Name != "Utilities2" || updated.Icon != "Zap" || updated.Color != "#f59e0b" || updated.IconColor != "#000000" {
		t.Fatalf("unexpected updated category: %+v", updated)
	}

	cats, err := s.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	var found bool
	for _, cat := range cats {
		if cat.ID == c.ID {
			found = true
			if cat.Icon != "Zap" || cat.Color != "#f59e0b" || cat.IconColor != "#000000" {
				t.Fatalf("listed category not updated: %+v", cat)
			}
		}
	}
	if !found {
		t.Fatal("category not found after update")
	}
}

func TestUpdateCategoryName(t *testing.T) {
	s := newStore(t)

	c, err := s.CreateCategory("Old Name", "Tag", "#6b7280", "#ffffff")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	other, err := s.CreateCategory("Taken Name", "Tag", "#6b7280", "#ffffff")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	updated, err := s.UpdateCategoryName(c.ID, "New Name")
	if err != nil {
		t.Fatalf("UpdateCategoryName: %v", err)
	}
	if updated.ID != c.ID || updated.Name != "New Name" || updated.Icon != "Tag" || updated.Color != "#6b7280" {
		t.Fatalf("unexpected updated category: %+v", updated)
	}

	catsAfter, err := s.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	var persisted bool
	for _, cat := range catsAfter {
		if cat.ID == c.ID && cat.Name == "New Name" {
			persisted = true
		}
	}
	if !persisted {
		t.Fatal("renamed category not found in list")
	}

	if _, err := s.UpdateCategoryName(c.ID, other.Name); err == nil {
		t.Fatal("want error renaming to an already-taken name, got nil")
	}
}

func TestDeleteTransaction(t *testing.T) {
	s := newStore(t)
	_, err := s.InsertTransactions([]model.Transaction{
		{AccountName: "Main", PartnerName: "LIDL", AmountEUR: -5, DedupeHash: "del-1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	txns, err := s.ListTransactions(0)
	if err != nil || len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d err=%v", len(txns), err)
	}
	id := txns[0].ID

	if err := s.DeleteTransaction(id); err != nil {
		t.Fatalf("DeleteTransaction: %v", err)
	}

	txns, err = s.ListTransactions(0)
	if err != nil || len(txns) != 0 {
		t.Fatalf("expected 0 transactions after delete, got %d err=%v", len(txns), err)
	}
}

func TestInsertTransactionsPersistsCategory(t *testing.T) {
	s := newStore(t)

	cats, err := s.ListCategories()
	if err != nil || len(cats) == 0 {
		t.Fatalf("expected seeded categories: %v", err)
	}
	catID := cats[0].ID

	n, err := s.InsertTransactions([]model.Transaction{
		{AccountName: "Main", PartnerName: "Imported Row", AmountEUR: -12, DedupeHash: "import-1",
			CategoryID: &catID, CategorizedBy: "import"},
	})
	if err != nil || n != 1 {
		t.Fatalf("insert: n=%d err=%v", n, err)
	}

	txns, err := s.ListTransactions(0)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	var found bool
	for _, tx := range txns {
		if tx.PartnerName == "Imported Row" {
			found = true
			if tx.CategoryName != cats[0].Name || tx.CategorizedBy != "import" {
				t.Fatalf("unexpected imported row: %+v", tx)
			}
		}
	}
	if !found {
		t.Fatal("imported row not found")
	}
}

func TestDeleteAccountCascadesTransactions(t *testing.T) {
	s := newStore(t)

	_, err := s.InsertTransactions([]model.Transaction{
		{AccountName: "ToDelete", PartnerName: "X", AmountEUR: -1, DedupeHash: "cascade-1"},
		{AccountName: "ToDelete", PartnerName: "Y", AmountEUR: -2, DedupeHash: "cascade-2"},
		{AccountName: "Keep", PartnerName: "Z", AmountEUR: -3, DedupeHash: "cascade-3"},
	})
	if err != nil {
		t.Fatal(err)
	}

	accounts, err := s.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	var toDeleteID int64
	for _, a := range accounts {
		if a.Name == "ToDelete" {
			toDeleteID = a.ID
		}
	}
	if toDeleteID == 0 {
		t.Fatal("ToDelete account not found")
	}

	if err := s.DeleteAccount(toDeleteID); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	accountsAfter, err := s.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range accountsAfter {
		if a.Name == "ToDelete" {
			t.Fatal("account still present after delete")
		}
	}

	txns, err := s.ListTransactions(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, tx := range txns {
		if tx.PartnerName == "X" || tx.PartnerName == "Y" {
			t.Fatalf("transaction from deleted account still present: %+v", tx)
		}
	}
	var keptFound bool
	for _, tx := range txns {
		if tx.PartnerName == "Z" {
			keptFound = true
		}
	}
	if !keptFound {
		t.Fatal("transaction from the other account should not have been deleted")
	}
}

func TestListCategoriesIncludesKind(t *testing.T) {
	s := newStore(t)

	cats, err := s.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	var found bool
	for _, c := range cats {
		if c.Kind != "income" && c.Kind != "expense" {
			t.Fatalf("category %q has invalid kind %q", c.Name, c.Kind)
		}
		if c.Name == "Groceries" {
			found = true
			if c.Kind != "expense" {
				t.Fatalf("want seeded Groceries to default to expense, got %q", c.Kind)
			}
		}
	}
	if !found {
		t.Fatal("Groceries category not found")
	}
}

func TestUpdateCategoryKind(t *testing.T) {
	s := newStore(t)

	c, err := s.CreateCategory("Freelance Income", "Wallet", "#16a34a", "#ffffff")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if c.Kind != "expense" {
		t.Fatalf("want new category to default to expense, got %q", c.Kind)
	}

	updated, err := s.UpdateCategoryKind(c.ID, "income")
	if err != nil {
		t.Fatalf("UpdateCategoryKind: %v", err)
	}
	if updated.Kind != "income" || updated.ID != c.ID || updated.Name != "Freelance Income" {
		t.Fatalf("unexpected updated category: %+v", updated)
	}

	cats, err := s.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	var persisted bool
	for _, cat := range cats {
		if cat.ID == c.ID {
			persisted = true
			if cat.Kind != "income" {
				t.Fatalf("kind update did not persist, got %q", cat.Kind)
			}
		}
	}
	if !persisted {
		t.Fatal("updated category not found in list")
	}
}
