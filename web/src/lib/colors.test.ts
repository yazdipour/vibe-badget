import { test } from "node:test";
import assert from "node:assert/strict";
import { PALETTE, readableTextColor } from "./colors.ts";

test("PALETTE has 10 distinct hex colors", () => {
  assert.equal(PALETTE.length, 10);
  assert.equal(new Set(PALETTE).size, 10);
  for (const c of PALETTE) {
    assert.match(c, /^#[0-9a-f]{6}$/i);
  }
});

test("readableTextColor: white background gets dark text", () => {
  assert.equal(readableTextColor("#ffffff"), "#1f2937");
});

test("readableTextColor: black background gets white text", () => {
  assert.equal(readableTextColor("#000000"), "#ffffff");
});
