import { test } from "node:test";
import assert from "node:assert/strict";
import { formatEUR } from "./format.ts";

test("formatEUR: adds thousands separators and 2 decimals", () => {
  assert.equal(formatEUR(29861.14), "29,861.14 €");
});

test("formatEUR: pads a whole number to 2 decimals", () => {
  assert.equal(formatEUR(1234.5), "1,234.50 €");
});

test("formatEUR: rounds to 2 decimals", () => {
  assert.equal(formatEUR(21423.880000000005), "21,423.88 €");
});

test("formatEUR: handles negative amounts", () => {
  assert.equal(formatEUR(-70), "-70.00 €");
});

test("formatEUR: handles zero", () => {
  assert.equal(formatEUR(0), "0.00 €");
});
