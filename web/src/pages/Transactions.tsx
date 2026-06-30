import { useEffect, useState } from "react";
import { api, type Account, type Tx } from "@/lib/api";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function Transactions() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [accountId, setAccountId] = useState<string>("all");
  const [rows, setRows] = useState<Tx[]>([]);

  useEffect(() => { api.accounts().then(setAccounts); }, []);
  useEffect(() => {
    api.transactions(accountId === "all" ? undefined : Number(accountId)).then(setRows);
  }, [accountId]);

  return (
    <div className="space-y-4">
      <Select value={accountId} onValueChange={(v) => setAccountId(v ?? "all")}>
        <SelectTrigger className="w-64"><SelectValue placeholder="Account" /></SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All accounts</SelectItem>
          {accounts.map((a) => <SelectItem key={a.id} value={String(a.id)}>{a.name}</SelectItem>)}
        </SelectContent>
      </Select>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Date</TableHead><TableHead>Partner</TableHead>
            <TableHead>Reference</TableHead><TableHead className="text-right">Amount</TableHead>
            <TableHead>Category</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((t) => (
            <TableRow key={t.id}>
              <TableCell>{t.booking_date}</TableCell>
              <TableCell>{t.partner_name}</TableCell>
              <TableCell className="text-muted-foreground">{t.payment_reference}</TableCell>
              <TableCell className={`text-right ${t.amount_eur < 0 ? "" : "text-green-600"}`}>
                {t.amount_eur.toFixed(2)} €
              </TableCell>
              <TableCell>
                {t.category_name
                  ? <Badge variant={t.categorized_by === "llm" ? "secondary" : "default"}>{t.category_name}</Badge>
                  : <span className="text-muted-foreground">—</span>}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
