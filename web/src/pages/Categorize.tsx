import { useEffect, useMemo, useState } from "react";
import { api, type Category, type Rule, type Tx, type CategorizeLogEntry } from "@/lib/api";
import { suggestRules, type Suggestion } from "@/lib/suggestions";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";

function suggestionKey(s: Pick<Suggestion, "partnerName" | "categoryName">): string {
  return `${s.partnerName} ${s.categoryName}`;
}

function sourceVariant(source: string): "default" | "secondary" | "outline" {
  if (source === "llm") return "secondary";
  if (source === "rule") return "default";
  return "outline";
}

export default function Categorize() {
  const [txns, setTxns] = useState<Tx[]>([]);
  const [rules, setRules] = useState<Rule[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [dismissed, setDismissed] = useState<Set<string>>(new Set());
  const [busy, setBusy] = useState(false);
  const [log, setLog] = useState<CategorizeLogEntry[] | null>(null);

  const reload = () => {
    api.transactions().then(setTxns);
    api.rules().then(setRules);
  };
  useEffect(() => { reload(); api.categories().then(setCategories); }, []);

  const suggestions = useMemo(
    () => suggestRules(txns, rules, categories).filter((s) => !dismissed.has(suggestionKey(s))),
    [txns, rules, categories, dismissed],
  );

  async function accept(s: Suggestion) {
    try {
      await api.createRule({ field: "partner_name", match_type: "exact", pattern: s.partnerName, category_id: s.categoryId });
      toast.success(`Rule created: ${s.partnerName} → ${s.categoryName}`);
      reload();
    } catch (e) {
      toast.error(String(e));
    }
  }

  function dismiss(s: Suggestion) {
    setDismissed((prev) => new Set(prev).add(suggestionKey(s)));
  }

  async function run() {
    setBusy(true);
    try {
      const r = await api.categorize();
      toast.success(`Rules: ${r.rules}, LLM: ${r.llm}, Skipped: ${r.skipped}`);
      setLog(r.log);
      reload();
    } catch (e) {
      toast.error(String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader><CardTitle>Suggested rules</CardTitle></CardHeader>
        <CardContent className="space-y-2">
          {suggestions.length === 0 ? (
            <p className="text-muted-foreground">No rule suggestions right now.</p>
          ) : (
            suggestions.map((s) => (
              <div key={suggestionKey(s)} className="flex items-center justify-between gap-2 rounded-lg border p-2">
                <span>
                  <strong>{s.partnerName}</strong> → {s.categoryName} (seen {s.count} times)
                </span>
                <div className="flex gap-2">
                  <Button size="sm" onClick={() => accept(s)}>Accept</Button>
                  <Button size="sm" variant="ghost" onClick={() => dismiss(s)}>Dismiss</Button>
                </div>
              </div>
            ))
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Run AI categorization</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <Button onClick={run} disabled={busy}>Run AI categorization</Button>

          {log && (
            log.length === 0 ? (
              <p className="text-muted-foreground">Nothing to categorize.</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Partner</TableHead><TableHead>Category</TableHead>
                    <TableHead>Source</TableHead><TableHead>Reason</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {log.map((entry) => (
                    <TableRow key={entry.tx_id}>
                      <TableCell>{entry.partner}</TableCell>
                      <TableCell>{entry.category || "—"}</TableCell>
                      <TableCell><Badge variant={sourceVariant(entry.source)}>{entry.source}</Badge></TableCell>
                      <TableCell className="text-muted-foreground">{entry.reason || "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )
          )}
        </CardContent>
      </Card>
    </div>
  );
}
