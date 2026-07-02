import { BrowserRouter, Routes, Route, NavLink } from "react-router-dom";
import { ThemeProvider } from "next-themes";
import {
  Wallet, Receipt, BarChart3, Tags, ListChecks, Settings as SettingsIcon,
} from "lucide-react";
import { buttonVariants } from "@/components/ui/button";
import { Toaster } from "@/components/ui/sonner";
import { cn } from "@/lib/utils";
import { AccountInfoDialog } from "@/components/AccountInfoDialog";
import Transactions from "./pages/Transactions";
import Rules from "./pages/Rules";
import Settings from "./pages/Settings";
import Visualization from "./pages/Visualization";
import Categorize from "./pages/Categorize";

const nav = [
  ["/", "Transactions", Receipt],
  ["/visualize", "Visualize", BarChart3],
  ["/categorize", "Categorize", Tags],
  ["/rules", "Rules", ListChecks],
] as const;

export default function App() {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <BrowserRouter>
        <div className="min-h-screen bg-background text-foreground">
          <header className="border-b">
            <nav className="mx-auto flex max-w-5xl items-center gap-4 p-4">
              <span className="flex items-center gap-2 font-bold">
                <Wallet size={18} />
                Vibe Badget
              </span>
              {nav.map(([to, label, Icon]) => (
                <NavLink key={to} to={to} end className={({ isActive }) =>
                  cn("flex items-center gap-1.5", isActive ? "font-medium underline" : "text-muted-foreground")}>
                  <Icon size={14} />
                  {label}
                </NavLink>
              ))}
              <div className="ml-auto flex items-center gap-1">
                <AccountInfoDialog />
                <NavLink
                  to="/settings"
                  className={({ isActive }) =>
                    cn(buttonVariants({ variant: "ghost", size: "icon" }), isActive && "bg-muted")}
                >
                  <SettingsIcon size={16} />
                </NavLink>
              </div>
            </nav>
          </header>
          <main className="mx-auto max-w-5xl p-4">
            <Routes>
              <Route path="/" element={<Transactions />} />
              <Route path="/visualize" element={<Visualization />} />
              <Route path="/categorize" element={<Categorize />} />
              <Route path="/rules" element={<Rules />} />
              <Route path="/settings" element={<Settings />} />
            </Routes>
          </main>
          <Toaster />
        </div>
      </BrowserRouter>
    </ThemeProvider>
  );
}
