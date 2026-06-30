package db

import "testing"

func TestOpenSeedsDefaults(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	var cats int
	if err := d.QueryRow(`SELECT count(*) FROM categories`).Scan(&cats); err != nil {
		t.Fatalf("count categories: %v", err)
	}
	if cats < 5 {
		t.Fatalf("want >=5 seeded categories, got %d", cats)
	}

	var lidl int
	err = d.QueryRow(`SELECT count(*) FROM rules WHERE pattern='Lidl'`).Scan(&lidl)
	if err != nil || lidl != 1 {
		t.Fatalf("want 1 Lidl rule, got %d err %v", lidl, err)
	}
}
