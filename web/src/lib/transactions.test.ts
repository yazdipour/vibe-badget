import { test } from "node:test";
import assert from "node:assert/strict";
import type { Tx } from "./api.ts";
import { filterTxns } from "./transactions.ts";

function mkTx(partial: Partial<Tx>): Tx {
  return {
    id: 1, account_id: 1, booking_date: "2026-01-01", partner_name: "Partner",
    partner_iban: "AT000", type: "Card Payment", payment_reference: "",
    amount_eur: -10, categorized_by: "", account_name: "Main", category_name: "",
    ...partial,
  };
}

test("filterTxns: empty search and 'all' category matches everything", () => {
  const txns = [mkTx({ id: 1 }), mkTx({ id: 2 })];
  assert.equal(filterTxns(txns, "", "all").length, 2);
});

test("filterTxns: search matches partner_name case-insensitively", () => {
  const txns = [
    mkTx({ id: 1, partner_name: "LIDL DANKT" }),
    mkTx({ id: 2, partner_name: "Some Cafe" }),
  ];
  const result = filterTxns(txns, "lidl", "all");
  assert.deepEqual(result.map((t) => t.id), [1]);
});

test("filterTxns: search matches payment_reference", () => {
  const txns = [
    mkTx({ id: 1, partner_name: "X", payment_reference: "invoice 4471" }),
    mkTx({ id: 2, partner_name: "Y", payment_reference: "rent" }),
  ];
  const result = filterTxns(txns, "4471", "all");
  assert.deepEqual(result.map((t) => t.id), [1]);
});

test("filterTxns: category filter 'uncategorized' matches empty category_name", () => {
  const txns = [
    mkTx({ id: 1, category_name: "" }),
    mkTx({ id: 2, category_name: "Groceries" }),
  ];
  const result = filterTxns(txns, "", "uncategorized");
  assert.deepEqual(result.map((t) => t.id), [1]);
});

test("filterTxns: category filter matches an exact category name", () => {
  const txns = [
    mkTx({ id: 1, category_name: "Groceries" }),
    mkTx({ id: 2, category_name: "Transport" }),
  ];
  const result = filterTxns(txns, "", "Groceries");
  assert.deepEqual(result.map((t) => t.id), [1]);
});

test("filterTxns: search and category filter compose (AND)", () => {
  const txns = [
    mkTx({ id: 1, partner_name: "LIDL", category_name: "Groceries" }),
    mkTx({ id: 2, partner_name: "LIDL", category_name: "Transport" }),
  ];
  const result = filterTxns(txns, "lidl", "Groceries");
  assert.deepEqual(result.map((t) => t.id), [1]);
});
