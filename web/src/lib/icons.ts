import * as Icons from "lucide-react";
import type { LucideIcon } from "lucide-react";

export const CATEGORY_ICONS = [
  "ShoppingCart", "Utensils", "Car", "Home", "Zap", "HeartPulse",
  "GraduationCap", "Plane", "Gift", "Wallet", "Landmark", "PiggyBank",
  "Gamepad2", "Shirt", "Wifi", "Tag",
] as const;

export function resolveIcon(name: string): LucideIcon {
  return (Icons as unknown as Record<string, LucideIcon>)[name] ?? Icons.Tag;
}
