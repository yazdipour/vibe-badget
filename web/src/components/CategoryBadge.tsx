import { resolveIcon } from "@/lib/icons";
import { Badge } from "@/components/ui/badge";
import type { Category } from "@/lib/api";

export function CategoryBadge({
  category, name, variant = "default",
}: {
  category: Category | undefined;
  name?: string;
  variant?: "default" | "secondary" | "outline";
}) {
  const Icon = resolveIcon(category?.icon ?? "Tag");
  const bg = category?.color ?? "#6b7280";
  const fg = category?.icon_color ?? "#ffffff";
  return (
    <Badge variant={variant} style={{ backgroundColor: bg, color: fg, borderColor: "transparent" }}>
      <Icon size={12} />
      {name ?? category?.name ?? "Unknown"}
    </Badge>
  );
}
