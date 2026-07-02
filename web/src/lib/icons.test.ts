import { test } from "node:test";
import assert from "node:assert/strict";
import { CATEGORY_ICONS, resolveIcon, suggestCategoryAppearance } from "./icons.ts";
import { Tag, Car } from "lucide-react";
import { PALETTE, readableTextColor } from "./colors.ts";

test("CATEGORY_ICONS includes Tag as the fallback/default", () => {
  assert.ok(CATEGORY_ICONS.includes("Tag"));
});

test("resolveIcon: returns the matching component for a known name", () => {
  assert.equal(resolveIcon("Car"), Car);
});

test("resolveIcon: falls back to Tag for an unknown name", () => {
  assert.equal(resolveIcon("NotARealIconName"), Tag);
});

test("suggestCategoryAppearance: chooses fitting icons and palette colors", () => {
  const grocery = suggestCategoryAppearance("Weekly groceries");
  assert.equal(grocery.icon, "ShoppingCart");
  assert.ok(PALETTE.includes(grocery.color));
  assert.equal(grocery.icon_color, readableTextColor(grocery.color));

  const unknown = suggestCategoryAppearance("Misc");
  assert.equal(unknown.icon, "Tag");
});
