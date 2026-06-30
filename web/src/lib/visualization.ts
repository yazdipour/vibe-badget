import type { Tx } from "./api.ts";

export type MonthlyTotals = { month: string; income: number; expenses: number };
export type CategorySlice = { name: string; value: number };
export type Summary = { income: number; expenses: number; net: number };

export function filterTransactions(
  txns: Tx[],
  accountId: string,
  from: string,
  to: string,
): Tx[] {
  return txns.filter((t) => {
    if (accountId !== "all" && String(t.account_id) !== accountId) return false;
    if (from && t.booking_date < from) return false;
    if (to && t.booking_date > to) return false;
    return true;
  });
}

export function summarize(txns: Tx[]): Summary {
  let income = 0;
  let expenses = 0;
  for (const t of txns) {
    if (t.amount_eur > 0) income += t.amount_eur;
    else expenses += t.amount_eur;
  }
  return { income, expenses, net: income + expenses };
}

export function monthlyTotals(txns: Tx[]): MonthlyTotals[] {
  const byMonth = new Map<string, MonthlyTotals>();
  for (const t of txns) {
    const month = t.booking_date.slice(0, 7);
    const bucket = byMonth.get(month) ?? { month, income: 0, expenses: 0 };
    if (t.amount_eur > 0) bucket.income += t.amount_eur;
    else bucket.expenses += -t.amount_eur;
    byMonth.set(month, bucket);
  }
  return [...byMonth.values()].sort((a, b) => a.month.localeCompare(b.month));
}

export function categoryTotals(txns: Tx[], sign: "income" | "expense"): CategorySlice[] {
  const filtered = txns.filter((t) =>
    sign === "income" ? t.amount_eur > 0 : t.amount_eur < 0,
  );
  const byCategory = new Map<string, number>();
  for (const t of filtered) {
    const name = t.category_name || "Uncategorized";
    byCategory.set(name, (byCategory.get(name) ?? 0) + Math.abs(t.amount_eur));
  }
  return [...byCategory.entries()]
    .map(([name, value]) => ({ name, value }))
    .sort((a, b) => b.value - a.value);
}
