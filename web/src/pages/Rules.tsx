import { useEffect, useState } from "react";
import { api, type Category, type Rule } from "@/lib/api";
import { downloadCsv } from "@/lib/csv";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";

const FIELDS = ["partner_iban", "partner_name", "type", "payment_reference"];
const MATCHES = ["exact", "keyword"];

export default function Rules() {
  const [rules, setRules] = useState<Rule[]>([]);
  const [cats, setCats] = useState<Category[]>([]);
  const [draft, setDraft] = useState({ field: "partner_name", match_type: "keyword", pattern: "", category_id: 0 });

  const reload = () => api.rules().then(setRules);
  useEffect(() => { reload(); api.categories().then(setCats); }, []);

  const catName = (id: number) => cats.find((c) => c.id === id)?.name ?? id;

  async function add() {
    if (!draft.pattern || !draft.category_id) { toast.error("pattern and category required"); return; }
    try { await api.createRule(draft); setDraft({ ...draft, pattern: "" }); reload(); }
    catch (e) { toast.error(String(e)); }
  }

  function exportRules() {
    const rows: string[][] = [["Field", "MatchType", "Pattern", "Category"]];
    for (const r of rules) {
      rows.push([r.field, r.match_type, r.pattern, String(catName(r.category_id))]);
    }
    downloadCsv("rules-export.csv", rows);
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end gap-2">
        <Select value={draft.field} onValueChange={(v) => setDraft({ ...draft, field: v ?? draft.field })}>
          <SelectTrigger className="w-44"><SelectValue /></SelectTrigger>
          <SelectContent>{FIELDS.map((f) => <SelectItem key={f} value={f}>{f}</SelectItem>)}</SelectContent>
        </Select>
        <Select value={draft.match_type} onValueChange={(v) => setDraft({ ...draft, match_type: v ?? draft.match_type })}>
          <SelectTrigger className="w-32"><SelectValue /></SelectTrigger>
          <SelectContent>{MATCHES.map((m) => <SelectItem key={m} value={m}>{m}</SelectItem>)}</SelectContent>
        </Select>
        <Input placeholder="pattern (e.g. Lidl)" value={draft.pattern}
          onChange={(e) => setDraft({ ...draft, pattern: e.target.value })} className="w-48" />
        <Select value={String(draft.category_id || "")} onValueChange={(v) => setDraft({ ...draft, category_id: Number(v) })}>
          <SelectTrigger className="w-44"><SelectValue placeholder="category" /></SelectTrigger>
          <SelectContent>{cats.map((c) => <SelectItem key={c.id} value={String(c.id)}>{c.name}</SelectItem>)}</SelectContent>
        </Select>
        <Button onClick={add}>Add rule</Button>
        <Button variant="outline" onClick={exportRules}>Export</Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow><TableHead>Field</TableHead><TableHead>Match</TableHead><TableHead>Pattern</TableHead>
            <TableHead>Category</TableHead><TableHead></TableHead></TableRow>
        </TableHeader>
        <TableBody>
          {rules.map((r) => (
            <TableRow key={r.id}>
              <TableCell>{r.field}</TableCell><TableCell>{r.match_type}</TableCell>
              <TableCell>{r.pattern}</TableCell><TableCell>{catName(r.category_id)}</TableCell>
              <TableCell className="text-right">
                <Button variant="ghost" size="sm"
                  onClick={async () => { await api.deleteRule(r.id); reload(); }}>Delete</Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
