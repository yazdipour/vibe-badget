import { test } from "node:test";
import assert from "node:assert/strict";
import { toCsvRow, parseCsv } from "./csv.ts";

test("toCsvRow: plain fields join with commas", () => {
  assert.equal(toCsvRow(["a", "b", "c"]), "a,b,c");
});

test("toCsvRow: quotes a field containing a comma", () => {
  assert.equal(toCsvRow(["Bills, Utilities", "x"]), '"Bills, Utilities",x');
});

test("toCsvRow: escapes embedded quotes by doubling them", () => {
  assert.equal(toCsvRow(['Say "hi"', "x"]), '"Say ""hi""",x');
});

test("parseCsv: simple multi-row input", () => {
  const rows = parseCsv("a,b,c\n1,2,3\n");
  assert.deepEqual(rows, [["a", "b", "c"], ["1", "2", "3"]]);
});

test("parseCsv: quoted field with embedded comma", () => {
  const rows = parseCsv('a,"b,c",d\n');
  assert.deepEqual(rows, [["a", "b,c", "d"]]);
});

test("parseCsv: quoted field with escaped quote", () => {
  const rows = parseCsv('a,"say ""hi""",c\n');
  assert.deepEqual(rows, [["a", 'say "hi"', "c"]]);
});

test("parseCsv: handles input with no trailing newline", () => {
  const rows = parseCsv("a,b\n1,2");
  assert.deepEqual(rows, [["a", "b"], ["1", "2"]]);
});

test("parseCsv and toCsvRow round-trip", () => {
  const original = ["Field, with comma", 'Quote " inside', "plain"];
  const encoded = toCsvRow(original);
  const [decoded] = parseCsv(encoded);
  assert.deepEqual(decoded, original);
});
