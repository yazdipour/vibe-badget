import { useEffect, useRef, useState } from "react";
import { api, type Category, type LLMHealth, type Rule } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ThemeToggle } from "@/components/ThemeToggle";
import { toast } from "sonner";

function healthVariant(status: string): "default" | "destructive" | "outline" {
  if (status === "ok") return "default";
  if (status === "unconfigured") return "outline";
  return "destructive";
}

type ConfigFile = {
  version: 1;
  categories: Pick<Category, "name" | "kind" | "icon" | "color" | "icon_color">[];
  rules: (Pick<Rule, "field" | "match_type" | "pattern"> & { category: string })[];
};

export default function Settings() {
  const [s, setS] = useState<Record<string, string>>({
    llm_base_url: "http://host.docker.internal:11434/v1",
    llm_model: "llama3.1",
    llm_concurrency: "4",
    llm_api_key: "",
  });
  const [health, setHealth] = useState<LLMHealth | null>(null);
  const [checking, setChecking] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => { api.getSettings().then((v) => setS((p) => ({ ...p, ...v, llm_api_key: "" }))); }, []);

  async function checkHealth() {
    setChecking(true);
    try {
      const h = await api.llmHealth();
      setHealth(h);
    } catch (e) {
      toast.error(String(e));
    } finally {
      setChecking(false);
    }
  }
  useEffect(() => { checkHealth(); }, []);

  async function save() {
    try {
      await api.putSettings(s);
      toast.success("Saved");
      checkHealth();
    } catch (e) {
      toast.error(String(e));
    }
  }

  async function loadConfigData(): Promise<[Rule[], Category[]]> {
    return Promise.all([api.rules(), api.categories()]);
  }

  async function exportConfig() {
    const [rules, cats] = await loadConfigData();
    const categoryById = new Map(cats.map((c) => [c.id, c.name]));
    const config: ConfigFile = {
      version: 1,
      categories: cats.map(({ name, kind, icon, color, icon_color }) => ({ name, kind, icon, color, icon_color })),
      rules: rules.map(({ field, match_type, pattern, category_id }) => ({
        field, match_type, pattern, category: categoryById.get(category_id) ?? String(category_id),
      })),
    };
    const blob = new Blob([JSON.stringify(config, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "vibe-wallet-config.json";
    a.click();
    URL.revokeObjectURL(url);
  }

  async function onImportFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      const config = JSON.parse(await file.text()) as ConfigFile;
      if (config.version !== 1 || !Array.isArray(config.categories) || !Array.isArray(config.rules)) {
        throw new Error("Unrecognized config format");
      }
      let cats = await api.categories();
      let categoriesImported = 0;
      for (const c of config.categories) {
        if (!cats.some((existing) => existing.name.toLowerCase() === c.name.toLowerCase())) {
          const created = await api.createCategory({ name: c.name, icon: c.icon, color: c.color, icon_color: c.icon_color });
          await api.updateCategoryKind(created.id, c.kind);
          categoriesImported++;
        }
      }
      cats = await api.categories();
      const catByName = new Map(cats.map((c) => [c.name.toLowerCase(), c]));
      let rulesImported = 0;
      let skipped = 0;
      for (const r of config.rules) {
        const cat = catByName.get(r.category.toLowerCase());
        if (!cat) {
          skipped++;
          continue;
        }
        try {
          await api.createRule({ field: r.field, match_type: r.match_type, pattern: r.pattern, category_id: cat.id });
          rulesImported++;
        } catch {
          skipped++;
        }
      }
      toast.success(`Imported ${categoriesImported} categories and ${rulesImported} rules, skipped ${skipped}`);
    } catch (err) {
      toast.error(String(err));
    } finally {
      e.target.value = "";
    }
  }

  const field = (key: string, label: string, placeholder = "") => (
    <label className="block space-y-1">
      <span className="text-sm text-muted-foreground">{label}</span>
      <Input value={s[key] ?? ""} placeholder={placeholder}
        onChange={(e) => setS({ ...s, [key]: e.target.value })} />
    </label>
  );

  return (
    <div className="max-w-lg space-y-4">
      <Card>
        <CardHeader><CardTitle>Appearance</CardTitle></CardHeader>
        <CardContent>
          <ThemeToggle />
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Rules and categories</CardTitle></CardHeader>
        <CardContent className="flex gap-2">
          <Button variant="outline" onClick={exportConfig}>Export config</Button>
          <Button variant="outline" onClick={() => fileInputRef.current?.click()}>Import config</Button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".json,application/json"
            className="hidden"
            onChange={onImportFile}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>LLM configuration (OpenAI-compatible / Ollama)</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          {field("llm_base_url", "Base URL", "http://host.docker.internal:11434/v1")}
          {field("llm_model", "Model", "llama3.1")}
          {field("llm_api_key", "API key (leave blank to keep current)")}
          {field("llm_concurrency", "Parallel workers")}
          <Button onClick={save}>Save</Button>

          <div className="flex items-center gap-2 border-t pt-4">
            {checking ? (
              <span className="text-sm text-muted-foreground">Checking…</span>
            ) : health ? (
              <>
                <Badge variant={healthVariant(health.status)}>{health.status}</Badge>
                <span className="text-sm text-muted-foreground">{health.message}</span>
              </>
            ) : null}
            <Button size="sm" variant="ghost" onClick={checkHealth} disabled={checking}>Recheck</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
