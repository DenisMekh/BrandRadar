import { useLocation, Link } from "react-router-dom";
import { Sun, Moon, Bell, ChevronRight, Menu, AlertTriangle, MessageSquare, Radio, Zap, XCircle, CheckCheck, ArrowRight } from "lucide-react";
import { useState, useEffect, useRef } from "react";
import { useBrands } from "@/hooks/use-brands";
import { useAlertHistory } from "@/hooks/use-alerts";
import { fmtRelative } from "@/lib/format";
import { cn } from "@/lib/utils";
import { useIsMobile } from "@/hooks/use-mobile";
import { clearCache } from "@/lib/api-cache";
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";

type NotifType = "spike" | "negative" | "collect_done" | "alert" | "collect_error" | "mention";

const notifConfig: Record<NotifType, { icon: typeof Bell; colorClass: string }> = {
  spike: { icon: Zap, colorClass: "text-destructive bg-destructive/15" },
  negative: { icon: AlertTriangle, colorClass: "text-destructive bg-destructive/15" },
  collect_done: { icon: Radio, colorClass: "text-success bg-success/15" },
  alert: { icon: Bell, colorClass: "text-destructive bg-destructive/15" },
  collect_error: { icon: XCircle, colorClass: "text-destructive bg-destructive/15" },
  mention: { icon: MessageSquare, colorClass: "text-primary bg-primary/15" },
};

interface Notification {
  id: string;
  type: NotifType;
  text: string;
  time: string;
  link: string;
  read: boolean;
}

function useTheme() {
  const [dark, setDark] = useState(() => {
    if (typeof window === "undefined") return true;
    const stored = localStorage.getItem("theme");
    if (stored === "light") return false;
    if (stored === "dark") return true;
    return true;
  });

  useEffect(() => {
    const root = document.documentElement;
    if (dark) {
      root.classList.add("dark");
      localStorage.setItem("theme", "dark");
    } else {
      root.classList.remove("dark");
      localStorage.setItem("theme", "light");
    }
  }, [dark]);

  return { dark, toggle: () => setDark((d) => !d) };
}

interface AppHeaderProps {
  onMenuClick: () => void;
}

const pageTitles: Record<string, string> = {
  "/": "Упоминания",
  "/brands": "Бренды",
  "/dashboard": "Дашборд",
  "/alerts": "Алерты",
  "/events": "Журнал событий",
  "/sources": "Источники",
};

function NotificationList({ notifications, onItemClick, onMarkAllRead, onClose, unreadCount }: {
  notifications: Notification[];
  onItemClick: (id: string) => void;
  onMarkAllRead: () => void;
  onClose: () => void;
  unreadCount: number;
}) {
  return (
    <>
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h4 className="text-sm font-medium text-foreground">Уведомления</h4>
          {unreadCount > 0 && (
            <span className="h-5 min-w-5 flex items-center justify-center rounded-full bg-destructive text-destructive-foreground text-[10px] font-bold px-1.5">
              {unreadCount}
            </span>
          )}
        </div>
        {unreadCount > 0 && (
          <button
            onClick={onMarkAllRead}
            className="text-xs text-primary hover:text-primary/80 flex items-center gap-1 transition-colors cursor-pointer"
          >
            <CheckCheck className="h-3.5 w-3.5" /> Прочитать все
          </button>
        )}
      </div>
      <div className="max-h-[400px] overflow-y-auto">
        {notifications.length === 0 ? (
          <div className="p-8 flex flex-col items-center justify-center text-center">
            <Bell className="h-8 w-8 text-muted-foreground/30 mb-2" />
            <p className="text-sm text-foreground font-medium">Нет новых уведомлений</p>
            <p className="text-xs text-muted-foreground mt-1">Здесь будут появляться алерты по вашим брендам</p>
          </div>
        ) : (
          notifications.map((n) => {
            const cfg = notifConfig[n.type];
            const Icon = cfg.icon;
            return (
              <Link
                key={n.id}
                to={n.link}
                onClick={() => onItemClick(n.id)}
                className={cn(
                  "flex items-start gap-3 px-4 py-3 hover:bg-secondary/50 transition-colors border-b border-border last:border-0",
                  !n.read && "border-l-2 border-l-primary bg-primary/5"
                )}
              >
                <div className={cn("w-8 h-8 rounded-full flex items-center justify-center shrink-0 mt-0.5", cfg.colorClass)}>
                  <Icon className="h-4 w-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className={cn("text-sm line-clamp-2", !n.read ? "text-foreground font-medium" : "text-foreground")}>{n.text}</p>
                  <p className="text-xs text-muted-foreground/60 mt-1">{fmtRelative(n.time)}</p>
                </div>
              </Link>
            );
          })
        )}
      </div>
      <div className="flex items-center justify-between px-4 py-2.5 border-t border-border">
        <Link
          to="/events"
          onClick={onClose}
          className="flex items-center gap-1.5 text-sm text-primary hover:text-primary/80 transition-colors"
        >
          Все события <ArrowRight className="h-3.5 w-3.5" />
        </Link>
      </div>
    </>
  );
}

export function AppHeader({ onMenuClick }: AppHeaderProps) {
  const location = useLocation();
  const { data: brands, isLoading: brandsLoading } = useBrands();
  const { data: realAlerts, isLoading: alertsLoading } = useAlertHistory(); // get global alerts without brandId filter

  const { dark, toggle: toggleTheme } = useTheme();
  const isMobile = useIsMobile();
  const [notifOpen, setNotifOpen] = useState(false);
  const notifRef = useRef<HTMLDivElement>(null);

  // 1. Last "Mark All as Read" timestamp
  const [lastReadTimestamp, setLastReadTimestamp] = useState<string>(() => {
    return localStorage.getItem("notifications_last_read_all") || "1970-01-01T00:00:00.000Z";
  });

  // 2. IDs of notifications marked as read INDIVIDUALLY after the timestamp
  const [readAlertIds, setReadAlertIds] = useState<Set<string>>(() => {
    try {
      const stored = localStorage.getItem("read_notifications");
      return stored ? new Set(JSON.parse(stored)) : new Set();
    } catch {
      return new Set();
    }
  });

  // Persist read status changes and timestamp to localStorage with aggressive error handling
  useEffect(() => {
    try {
      localStorage.setItem("notifications_last_read_all", lastReadTimestamp);
      
      // Limit individual IDs to 20 to save maximum space
      const idsArray = Array.from(readAlertIds).slice(-20);
      localStorage.setItem("read_notifications", JSON.stringify(idsArray));
    } catch (e) {
      console.warn("Storage critically full. Attempting emergency cleanup...", e);
      try {
        // Emergency: clear the large API cache to make room for vital UI state
        clearCache();
        // Also clear our own ID list just in case
        localStorage.removeItem("read_notifications");
        localStorage.setItem("notifications_last_read_all", lastReadTimestamp);
      } catch (innerE) {
        console.error("Storage remains full even after cache cleanup. Site data reset may be required.", innerE);
      }
    }
  }, [readAlertIds, lastReadTimestamp]);

  // Map real alerts to the Notification interface required by the UI
  const notifications: Notification[] = (realAlerts || [])
    .slice(0, 15) // take top 15 most recent globally
    .map(a => {
      // Create a readable label for the brand internally here using brands map if available, or just fallback
      let brandName = "Неизвестный бренд";
      if (a.brand_id && brands) {
        const b = brands.find(br => br.id === a.brand_id);
        if (b) brandName = b.name;
      }

      const windowMinutes = a.window_start && a.window_end
        ? Math.round((new Date(a.window_end).getTime() - new Date(a.window_start).getTime()) / 60000)
        : undefined;

      const title = a.mentions_count === 0
        ? `Алерт сработал по бренду «${brandName}»${windowMinutes ? ` (окно: ${windowMinutes} мин)` : ""}`
        : `Всплеск упоминаний: найдено ${a.mentions_count} совпадений по бренду «${brandName}»${windowMinutes ? ` (окно: ${windowMinutes} мин)` : ""}`;

      const firedAt = a.fired_at || new Date().toISOString();
      const isReadByTimestamp = new Date(firedAt) <= new Date(lastReadTimestamp);
      const isReadManually = readAlertIds.has(a.id || "");

      return {
        id: a.id || Math.random().toString(),
        type: "spike" as NotifType,
        text: title,
        time: firedAt,
        link: "/alerts",
        read: isReadByTimestamp || isReadManually
      };
    });

  const unreadCount = notifications.filter((n) => !n.read).length;

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) setNotifOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const markAllRead = () => {
    // Set timestamp to now. This makes everything currently existing (and older) read.
    const now = new Date().toISOString();
    setLastReadTimestamp(now);
    // Clear individual IDs as they are now covered by the timestamp
    setReadAlertIds(new Set());
  };

  const handleNotifClick = (id: string) => {
    const newReads = new Set(readAlertIds);
    newReads.add(id);
    setReadAlertIds(newReads);
    setNotifOpen(false);
  };

  const title = pageTitles[location.pathname] || "Страница";

  const breadcrumbs: { label: string; href?: string }[] = [];
  const pathParts = location.pathname.split("/").filter(Boolean);

  if (pathParts[0] === "brands" && pathParts[1]) {
    const brand = brands?.find((b) => b.id === pathParts[1]);
    breadcrumbs.push({ label: "Бренды", href: "/brands" });
    breadcrumbs.push({ label: brand?.name ?? "Бренд" });
  }

  const searchParams = new URLSearchParams(location.search);
  if (location.pathname === "/mentions" && searchParams.get("brand_id")) {
    const brand = brands?.find((b) => b.id === searchParams.get("brand_id"));
    if (brand) {
      breadcrumbs.push({ label: "Упоминания", href: "/mentions" });
      breadcrumbs.push({ label: brand.name });
    }
  }

  return (
    <header className="h-14 md:h-16 shrink-0 border-b border-border flex items-center justify-between px-4 md:px-6 bg-background">
      <div className="flex items-center gap-2 min-w-0">
        <button
          onClick={onMenuClick}
          className="md:hidden p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors cursor-pointer min-w-[44px] min-h-[44px] flex items-center justify-center"
          aria-label="Открыть меню"
        >
          <Menu className="h-5 w-5" />
        </button>

        {breadcrumbs.length > 0 ? (
          <nav className="flex items-center gap-1 text-sm min-w-0">
            {breadcrumbs.map((bc, i) => (
              <span key={i} className="flex items-center gap-1 min-w-0">
                {i > 0 && <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/50 shrink-0" />}
                {bc.href ? (
                  <Link to={bc.href} className="text-muted-foreground hover:text-foreground transition-colors">{bc.label}</Link>
                ) : (
                  <span className="text-foreground font-medium truncate">{bc.label}</span>
                )}
              </span>
            ))}
          </nav>
        ) : (
          <h2 className="text-lg md:text-xl font-semibold text-foreground truncate">{title}</h2>
        )}
      </div>

      <div className="flex items-center gap-1">
        {/* Notifications */}
        <div className="relative" ref={notifRef}>
          <button
            onClick={() => setNotifOpen(!notifOpen)}
            className="relative p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors cursor-pointer min-w-[44px] min-h-[44px] flex items-center justify-center"
            aria-label="Уведомления"
          >
            {/* Show an active animation dot if unread, or nothing if caught up */}
            <Bell className={cn("h-5 w-5", unreadCount > 0 && "text-foreground")} />
            {unreadCount > 0 && (
              <span className="absolute top-1.5 right-1.5 w-2.5 h-2.5 bg-destructive rounded-full animate-pulse" />
            )}
          </button>

          {/* Desktop dropdown */}
          {notifOpen && !isMobile && (
            <div className="absolute right-0 top-full mt-2 w-[360px] bg-card border border-border rounded-xl shadow-xl z-50 overflow-hidden">
              <NotificationList
                notifications={notifications}
                onItemClick={handleNotifClick}
                onMarkAllRead={markAllRead}
                onClose={() => setNotifOpen(false)}
                unreadCount={unreadCount}
              />
            </div>
          )}
        </div>

        {/* Mobile bottom sheet for notifications */}
        {isMobile && (
          <Drawer open={notifOpen} onOpenChange={setNotifOpen}>
            <DrawerContent className="max-h-[85vh]">
              <DrawerHeader className="sr-only">
                <DrawerTitle>Уведомления</DrawerTitle>
              </DrawerHeader>
              <NotificationList
                notifications={notifications}
                onItemClick={handleNotifClick}
                onMarkAllRead={markAllRead}
                onClose={() => setNotifOpen(false)}
                unreadCount={unreadCount}
              />
            </DrawerContent>
          </Drawer>
        )}

        {/* Theme toggle */}
        <button
          onClick={toggleTheme}
          className="p-2.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors cursor-pointer min-w-[44px] min-h-[44px] flex items-center justify-center"
          aria-label="Переключить тему"
        >
          {dark ? <Moon className="h-5 w-5" /> : <Sun className="h-5 w-5" />}
        </button>
      </div>
    </header>
  );
}
