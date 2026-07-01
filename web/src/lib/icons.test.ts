import { test } from "node:test";
import assert from "node:assert/strict";
import { CATEGORY_ICONS, resolveIcon } from "./icons.ts";
import { Tag, Car } from "lucide-react";

test("CATEGORY_ICONS includes Tag as the fallback/default", () => {
  assert.ok(CATEGORY_ICONS.includes("Tag"));
});

test("resolveIcon: returns the matching component for a known name", () => {
  assert.equal(resolveIcon("Car"), Car);
});

test("resolveIcon: falls back to Tag for an unknown name", () => {
  assert.equal(resolveIcon("NotARealIconName"), Tag);
});
