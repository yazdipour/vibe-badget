import { useState } from "react";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";

export default function Upload() {
  const [busy, setBusy] = useState(false);

  async function onUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setBusy(true);
    try {
      const { inserted } = await api.upload(file);
      toast.success(`Imported ${inserted} new transactions`);
    } catch (err) {
      toast.error(String(err));
    } finally {
      setBusy(false);
      e.target.value = "";
    }
  }

  async function onCategorize() {
    setBusy(true);
    try {
      const r = await api.categorize();
      toast.success(`Rules: ${r.rules}, LLM: ${r.llm}, Skipped: ${r.skipped}`);
    } catch (err) {
      toast.error(String(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <Card>
      <CardHeader><CardTitle>Import transactions</CardTitle></CardHeader>
      <CardContent className="space-y-4">
        <Input type="file" accept=".csv" onChange={onUpload} disabled={busy} />
        <Button onClick={onCategorize} disabled={busy}>Run categorisation</Button>
      </CardContent>
    </Card>
  );
}
