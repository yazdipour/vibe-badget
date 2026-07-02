import { BrowserRouter, Routes, Route, NavLink, useLocation } from "react-router-dom";
import { ThemeProvider } from "next-themes";
import {
  Wallet, Receipt, BarChart3, Tags, ListChecks, Settings as SettingsIcon,
} from "lucide-react";
import {
  SidebarProvider, Sidebar, SidebarHeader, SidebarContent, SidebarFooter,
  SidebarMenu, SidebarMenuItem, SidebarMenuButton, SidebarGroup, SidebarGroupContent,
  SidebarInset, SidebarTrigger,
} from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import { Toaster } from "@/components/ui/sonner";
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

function currentTitle(pathname: string): string {
  if (pathname === "/settings") return "Settings";
  const match = nav.find(([to]) => (to === "/" ? pathname === "/" : pathname.startsWith(to)));
  return match?.[1] ?? "Vibe Badget";
}

function AppSidebar() {
  const location = useLocation();
  return (
    <Sidebar collapsible="icon" variant="inset">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton className="data-[slot=sidebar-menu-button]:p-1.5!" render={<NavLink to="/" end />}>
              <Wallet className="size-5!" />
              <span className="text-base font-semibold">Vibe Badget</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent className="flex flex-col gap-2">
            <SidebarMenu>
              {nav.map(([to, label, Icon]) => (
                <SidebarMenuItem key={to}>
                  <SidebarMenuButton
                    render={<NavLink to={to} end />}
                    isActive={to === "/" ? location.pathname === "/" : location.pathname.startsWith(to)}
                    tooltip={label}
                  >
                    <Icon />
                    <span>{label}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              render={<NavLink to="/settings" />}
              isActive={location.pathname === "/settings"}
              tooltip="Settings"
            >
              <SettingsIcon />
              <span>Settings</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
}

function SiteHeader() {
  const location = useLocation();
  return (
    <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-(--header-height)">
      <div className="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator orientation="vertical" className="mx-2 h-4 data-vertical:self-auto" />
        <h1 className="text-base font-medium">{currentTitle(location.pathname)}</h1>
        <div className="ml-auto">
          <AccountInfoDialog />
        </div>
      </div>
    </header>
  );
}

export default function App() {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <BrowserRouter>
        <SidebarProvider
          style={{
            "--sidebar-width": "calc(var(--spacing) * 64)",
            "--header-height": "calc(var(--spacing) * 12)",
          } as React.CSSProperties}
        >
          <AppSidebar />
          <SidebarInset>
            <SiteHeader />
            <div className="flex flex-1 flex-col">
              <div className="@container/main flex flex-1 flex-col gap-2">
                <div className="flex flex-col gap-4 px-4 py-4 md:gap-6 md:py-6 lg:px-6">
                  <Routes>
                    <Route path="/" element={<Transactions />} />
                    <Route path="/visualize" element={<Visualization />} />
                    <Route path="/categorize" element={<Categorize />} />
                    <Route path="/rules" element={<Rules />} />
                    <Route path="/settings" element={<Settings />} />
                  </Routes>
                </div>
              </div>
            </div>
          </SidebarInset>
        </SidebarProvider>
        <Toaster />
      </BrowserRouter>
    </ThemeProvider>
  );
}
