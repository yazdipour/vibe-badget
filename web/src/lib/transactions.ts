import type { Tx } from "./api.ts";

export function filterTxns(txns: Tx[], search: string, categoryFilter: string): Tx[] {
  const needle = search.trim().toLowerCase();
  return txns.filter((t) => {
    if (needle) {
      const haystack = `${t.partner_name} ${t.payment_reference}`.toLowerCase();
      if (!haystack.includes(needle)) return false;
    }
    if (categoryFilter === "uncategorized") {
      if (t.category_name) return false;
    } else if (categoryFilter !== "all") {
      if (t.category_name !== categoryFilter) return false;
    }
    return true;
  });
}
