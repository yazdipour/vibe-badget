import * as Icons from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { PALETTE, readableTextColor } from "@/lib/colors";

export const CATEGORY_ICONS = [
  "ShoppingCart", "Utensils", "Car", "Home", "Zap", "HeartPulse",
  "GraduationCap", "Plane", "Gift", "Wallet", "Landmark", "PiggyBank",
  "Gamepad2", "Shirt", "Wifi", "Tag",
] as const;

export function resolveIcon(name: string): LucideIcon {
  return (Icons as unknown as Record<string, LucideIcon>)[name] ?? Icons.Tag;
}

const ICON_KEYWORDS: [string[], (typeof CATEGORY_ICONS)[number]][] = [
  [["grocery", "groceries", "supermarket", "market", "food"], "ShoppingCart"],
  [["restaurant", "eating", "dining", "cafe", "coffee", "lunch", "dinner"], "Utensils"],
  [["transport", "car", "fuel", "gas", "taxi", "train", "bus", "metro"], "Car"],
  [["rent", "home", "house", "mortgage"], "Home"],
  [["bill", "utility", "utilities", "electric", "energy", "power"], "Zap"],
  [["health", "doctor", "medical", "pharmacy", "fitness"], "HeartPulse"],
  [["school", "education", "course", "book", "university"], "GraduationCap"],
  [["travel", "flight", "hotel", "vacation", "trip"], "Plane"],
  [["gift", "present", "donation"], "Gift"],
  [["income", "salary", "payroll", "wage"], "Wallet"],
  [["bank", "tax", "fee", "loan"], "Landmark"],
  [["saving", "savings", "investment", "invest"], "PiggyBank"],
  [["entertainment", "game", "games", "movie", "cinema"], "Gamepad2"],
  [["shopping", "clothes", "clothing", "fashion"], "Shirt"],
  [["internet", "wifi", "phone", "mobile"], "Wifi"],
];

export function suggestCategoryAppearance(name: string): { icon: string; color: string; icon_color: string } {
  const normalized = name.toLowerCase();
  const match = ICON_KEYWORDS.find(([keywords]) => keywords.some((keyword) => normalized.includes(keyword)));
  const icon = match?.[1] ?? "Tag";
  const color = PALETTE[Math.floor(Math.random() * PALETTE.length)];
  return { icon, color, icon_color: readableTextColor(color) };
}
