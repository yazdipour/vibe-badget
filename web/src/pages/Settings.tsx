import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";

export default function Settings() {
  const [s, setS] = useState<Record<string, string>>({
    llm_base_url: "http://host.docker.internal:11434/v1",
    llm_model: "llama3.1",
    llm_concurrency: "4",
    llm_api_key: "",
  });

  useEffect(() => { api.getSettings().then((v) => setS((p) => ({ ...p, ...v, llm_api_key: "" }))); }, []);

  async function save() {
    try { await api.putSettings(s); toast.success("Saved"); }
    catch (e) { toast.error(String(e)); }
  }

  const field = (key: string, label: string, placeholder = "") => (
    <label className="block space-y-1">
      <span className="text-sm text-muted-foreground">{label}</span>
      <Input value={s[key] ?? ""} placeholder={placeholder}
        onChange={(e) => setS({ ...s, [key]: e.target.value })} />
    </label>
  );

  return (
    <Card className="max-w-lg">
      <CardHeader><CardTitle>LLM configuration (OpenAI-compatible / Ollama)</CardTitle></CardHeader>
      <CardContent className="space-y-4">
        {field("llm_base_url", "Base URL", "http://host.docker.internal:11434/v1")}
        {field("llm_model", "Model", "llama3.1")}
        {field("llm_api_key", "API key (leave blank to keep current)")}
        {field("llm_concurrency", "Parallel workers")}
        <Button onClick={save}>Save</Button>
      </CardContent>
    </Card>
  );
}
