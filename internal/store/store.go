package store

import (
	"database/sql"

	"github.com/sh-yazdipour/vibe-badget/internal/model"
)

type Store struct{ db *sql.DB }

func New(d *sql.DB) *Store { return &Store{db: d} }

func (s *Store) UpsertAccount(name string) (int64, error) {
	if _, err := s.db.Exec(`INSERT OR IGNORE INTO accounts(name) VALUES(?)`, name); err != nil {
		return 0, err
	}
	var id int64
	err := s.db.QueryRow(`SELECT id FROM accounts WHERE name=?`, name).Scan(&id)
	return id, err
}

func (s *Store) InsertTransactions(txns []model.Transaction) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	accIDs := map[string]int64{}
	inserted := 0
	for _, t := range txns {
		id, ok := accIDs[t.AccountName]
		if !ok {
			if err := tx.QueryRow(`INSERT INTO accounts(name) VALUES(?)
				ON CONFLICT(name) DO UPDATE SET name=excluded.name RETURNING id`, t.AccountName).Scan(&id); err != nil {
				return 0, err
			}
			accIDs[t.AccountName] = id
		}
		res, err := tx.Exec(`INSERT OR IGNORE INTO transactions
			(account_id,booking_date,value_date,partner_name,partner_iban,type,
			 payment_reference,amount_eur,original_amount,original_currency,exchange_rate,dedupe_hash)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
			id, t.BookingDate, t.ValueDate, t.PartnerName, t.PartnerIban, t.Type,
			t.PaymentReference, t.AmountEUR, t.OriginalAmount, t.OriginalCurrency, t.ExchangeRate, t.DedupeHash)
		if err != nil {
			return 0, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	}
	return inserted, tx.Commit()
}
