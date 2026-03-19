import { Link, useLocation } from "react-router-dom";
import {
  LayoutDashboard,
  MessageSquare,
  Bell,
  ScrollText,
  Globe,
  X,
  Tag,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { HealthIndicator } from "@/components/HealthIndicator";
import { useAlertHistory } from "@/hooks/use-alerts";

const navItems = [
  { title: "Упоминания", path: "/", icon: MessageSquare, badge: null },
  { title: "Бренды", path: "/brands", icon: Tag, badge: null },
  { title: "Дашборд", path: "/dashboard", icon: LayoutDashboard, badge: null },
  { title: "Алерты", path: "/alerts", icon: Bell, badge: "3" },
  { title: "Журнал событий", path: "/events", icon: ScrollText, badge: null },
  { title: "Источники", path: "/sources", icon: Globe, badge: null },
];

interface AppSidebarProps {
  mobileOpen: boolean;
  onMobileClose: () => void;
}

export function AppSidebar({ mobileOpen, onMobileClose }: AppSidebarProps) {
  const location = useLocation();
  const { data: alerts = [] } = useAlertHistory();

  const lastViewed = localStorage.getItem("lastViewedAlerts");
  const lastReadAll = localStorage.getItem("notifications_last_read_all");
  const readAlertIdsStr = localStorage.getItem("read_notifications");
  let readAlertIds = new Set<string>();
  try {
    if (readAlertIdsStr) readAlertIds = new Set(JSON.parse(readAlertIdsStr));
  } catch (e) { /* ignore */ }

  const unreadAlerts = alerts.filter(a => {
    // If already marked as read in the header/all-read, it's not unread here
    if (a.id && readAlertIds.has(a.id)) return false;
    
    // Check against "Mark All as Read" timestamp
    const firedAtStr = a.fired_at || a.created_at;
    if (firedAtStr) {
      const firedAt = new Date(firedAtStr);
      if (lastReadAll && firedAt <= new Date(lastReadAll)) return false;
      if (lastViewed && firedAt <= new Date(lastViewed)) return false;
    }

    // Otherwise check timestamp
    if (!lastViewed && !lastReadAll) return true;
    return true; // If we reached here, it's newer than both
  });

  const badgeValue = unreadAlerts.length > 9 ? "9+" : unreadAlerts.length > 0 ? String(unreadAlerts.length) : null;

  const dynamicNavItems = navItems.map(item =>
    item.path === "/alerts" ? { ...item, badge: badgeValue } : item
  );

  const isActive = (path: string) => location.pathname === path || (path !== "/" && location.pathname.startsWith(path));

  const navContent = (
    <>
      <div className="px-4 py-5 border-b border-border">
        <Link to="/" onClick={onMobileClose} className="inline-block">
          <h1 className="text-xl font-bold bg-gradient-to-r from-violet-500 to-fuchsia-500 bg-clip-text text-transparent">BrandRadar</h1>
        </Link>
      </div>
      <nav className="flex-1 px-3 py-4 space-y-1">
        {dynamicNavItems.map((item) => {
          const active = isActive(item.path);
          return (
            <Link key={item.path} to={item.path} onClick={onMobileClose}
              className={cn("group flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 relative cursor-pointer min-h-[44px]",
                "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
                active ? "bg-primary/20 text-primary" : "text-muted-foreground hover:bg-secondary/50 hover:text-foreground")}>
              {active && <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 bg-primary rounded-r" />}
              <item.icon className="h-4 w-4 shrink-0" />
              <span className="flex-1">{item.title}</span>
              {item.badge && (
                <span className={cn("h-5 min-w-5 items-center justify-center rounded-full bg-destructive text-destructive-foreground text-[10px] font-bold px-1.5 inline-flex",
                  Number(item.badge) > 0 && "animate-pulse")}>{item.badge}</span>
              )}
            </Link>
          );
        })}
      </nav>
      <div className="px-5 py-4 border-t border-border flex items-center justify-between">
        <HealthIndicator />
        <span className="bg-secondary text-muted-foreground text-xs rounded px-1.5 py-0.5 font-mono">⌘K</span>
      </div>
    </>
  );

  return (
    <>
      {/* Mobile drawer overlay */}
      {mobileOpen && (
        <div
          className="md:hidden fixed inset-0 z-40 bg-black/60 backdrop-blur-sm"
          onClick={onMobileClose}
        />
      )}

      {/* Mobile drawer */}
      <aside
        className={cn(
          "md:hidden fixed inset-y-0 left-0 z-50 w-[280px] bg-card border-r border-border flex flex-col transition-transform duration-300 ease-out",
          mobileOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        <button
          onClick={onMobileClose}
          className="absolute top-4 right-4 p-2 text-muted-foreground hover:text-foreground cursor-pointer transition-colors min-w-[44px] min-h-[44px] flex items-center justify-center"
          aria-label="Закрыть меню"
        >
          <X className="h-5 w-5" />
        </button>
        {navContent}
      </aside>

      {/* Tablet: icons only */}
      <aside className="hidden md:flex lg:hidden fixed inset-y-0 left-0 w-16 z-30 border-r border-border bg-card flex-col items-center h-screen overflow-y-auto">
        <div className="py-5">
          <Link to="/">
            <span className="text-lg font-bold bg-clip-text text-transparent" style={{ backgroundImage: "var(--gradient-brand)" }}>B</span>
          </Link>
        </div>
        <div className="mx-2 border-t border-border w-8" />
        <nav className="flex-1 py-4 space-y-1 flex flex-col items-center">
          {dynamicNavItems.map((item) => {
            const active = isActive(item.path);
            return (
              <Tooltip key={item.path} delayDuration={0}>
                <TooltipTrigger asChild>
                  <Link to={item.path}
                    className={cn("relative flex items-center justify-center w-11 h-11 rounded-lg transition-all duration-200 cursor-pointer",
                      "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
                      active ? "bg-primary/20 text-primary" : "text-muted-foreground hover:bg-secondary/50 hover:text-foreground")}>
                    {active && <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 bg-primary rounded-r" />}
                    <item.icon className="h-4 w-4" />
                    {item.badge && (
                      <span className={cn("absolute -top-0.5 -right-0.5 h-4 min-w-4 flex items-center justify-center rounded-full bg-destructive text-destructive-foreground text-[9px] font-bold px-1",
                        Number(item.badge) > 0 && "animate-pulse")}>{item.badge}</span>
                    )}
                  </Link>
                </TooltipTrigger>
                <TooltipContent side="right" className="text-xs">{item.title}</TooltipContent>
              </Tooltip>
            );
          })}
        </nav>
        <div className="py-4"><HealthIndicator compact /></div>
      </aside>

      {/* Desktop: full sidebar */}
      <aside className="hidden lg:flex fixed inset-y-0 left-0 w-60 z-30 border-r border-border bg-card flex-col h-screen overflow-y-auto">{navContent}</aside>
    </>
  );
}
