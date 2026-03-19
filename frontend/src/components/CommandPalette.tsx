import { useEffect, useState, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import {
  LayoutDashboard, Tag, MessageSquare, Bell, ScrollText, Globe, Plus,
  TrendingDown, CheckCircle, Moon, Sun, Search, SearchX,
} from "lucide-react";
import { useBrands } from "@/hooks/use-brands";

interface PaletteItem {
  id: string;
  label: string;
  icon: typeof Search;
  group: string;
  hint: string;
  action: () => void;
  keywords?: string[];
  chips?: string[];
}

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const [darkMode, setDarkMode] = useState(() => document.documentElement.classList.contains("dark"));
  const navigate = useNavigate();
  const brandsQuery = useBrands();

  const close = useCallback(() => {
    setOpen(false);
    setQuery("");
    setActiveIndex(0);
  }, []);

  const go = useCallback((path: string) => {
    close();
    navigate(path);
  }, [close, navigate]);

  const toggleTheme = useCallback(() => {
    const next = !darkMode;
    setDarkMode(next);
    document.documentElement.classList.toggle("dark", next);
    localStorage.setItem("theme", next ? "dark" : "light");
    close();
  }, [darkMode, close]);

  const staticItems = useMemo<PaletteItem[]>(() => [
    { id: "nav-dash", label: "Дашборд", icon: LayoutDashboard, group: "Навигация", hint: "Навигация", action: () => go("/") },
    { id: "nav-brands", label: "Бренды", icon: Tag, group: "Навигация", hint: "Навигация", action: () => go("/brands") },
    { id: "nav-mentions", label: "Упоминания", icon: MessageSquare, group: "Навигация", hint: "Навигация", action: () => go("/mentions") },
    { id: "nav-alerts", label: "Алерты", icon: Bell, group: "Навигация", hint: "Навигация", action: () => go("/alerts") },
    { id: "nav-events", label: "Журнал событий", icon: ScrollText, group: "Навигация", hint: "Навигация", action: () => go("/events") },
    { id: "nav-sources", label: "Источники", icon: Globe, group: "Навигация", hint: "Навигация", action: () => go("/sources") },
    { id: "act-negative", label: "Показать негатив", icon: TrendingDown, group: "Быстрые действия", hint: "Действие", action: () => go("/mentions?sentiment=negative") },
    { id: "act-relevant", label: "Показать релевантные", icon: CheckCircle, group: "Быстрые действия", hint: "Действие", action: () => go("/mentions?is_relevant=true") },
    { id: "act-brand", label: "Добавить бренд", icon: Plus, group: "Быстрые действия", hint: "Действие", action: () => go("/brands?action=create") },
    { id: "act-theme", label: "Сменить тему", icon: darkMode ? Sun : Moon, group: "Быстрые действия", hint: "Действие", action: toggleTheme },
  ], [go, toggleTheme, darkMode]);

  const brandItems = useMemo<PaletteItem[]>(() => {
    if (!brandsQuery.data) return [];
    return brandsQuery.data.map((b) => ({
      id: `brand-${b.id}`,
      label: b.name,
      icon: Tag,
      group: "Бренды",
      hint: "Бренд",
      action: () => go(`/brands/${b.id}`),
      keywords: b.keywords,
      chips: b.keywords?.slice(0, 3),
    }));
  }, [brandsQuery.data, go]);

  const allItems = useMemo(() => [...staticItems, ...brandItems], [staticItems, brandItems]);

  const filtered = useMemo(() => {
    if (!query.trim()) return staticItems; // hide brands when no query
    const q = query.toLowerCase();
    return allItems.filter((item) => {
      if (item.label.toLowerCase().includes(q)) return true;
      if (item.keywords?.some((k) => k.toLowerCase().includes(q))) return true;
      return false;
    });
  }, [query, staticItems, allItems]);

  // Group filtered results
  const grouped = useMemo(() => {
    const groups: { name: string; items: PaletteItem[] }[] = [];
    const order = ["Навигация", "Быстрые действия", "Бренды"];
    for (const name of order) {
      const items = filtered.filter((i) => i.group === name);
      if (items.length > 0) groups.push({ name, items });
    }
    return groups;
  }, [filtered]);

  const flatFiltered = useMemo(() => grouped.flatMap((g) => g.items), [grouped]);

  // Reset active index on filter change
  useEffect(() => { setActiveIndex(0); }, [query]);

  // Keyboard shortcut to open
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, []);

  // Keyboard navigation inside palette
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveIndex((i) => (i + 1) % flatFiltered.length);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveIndex((i) => (i - 1 + flatFiltered.length) % flatFiltered.length);
    } else if (e.key === "Enter") {
      e.preventDefault();
      flatFiltered[activeIndex]?.action();
    } else if (e.key === "Escape") {
      e.preventDefault();
      close();
    }
  }, [flatFiltered, activeIndex, close]);

  if (!open) return null;

  let itemCounter = -1;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]" onClick={close}>
      {/* Overlay */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />

      {/* Dialog */}
      <div
        className="relative w-[560px] max-w-[calc(100vw-2rem)] max-h-[420px] bg-card border border-border rounded-xl shadow-2xl overflow-hidden animate-scale-in"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Search input */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-border">
          <Search className="h-5 w-5 text-muted-foreground shrink-0" />
          <input
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Искать команды, бренды, страницы..."
            className="flex-1 bg-transparent text-base text-foreground placeholder:text-muted-foreground focus:outline-none"
          />
          <kbd className="hidden sm:inline-flex text-xs text-muted-foreground/60 bg-secondary px-1.5 py-0.5 rounded">Esc</kbd>
        </div>

        {/* Results */}
        <div className="overflow-y-auto max-h-[360px] py-2">
          {flatFiltered.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 gap-2">
              <SearchX className="h-8 w-8 text-muted-foreground/40" />
              <p className="text-sm text-muted-foreground">Ничего не найдено</p>
            </div>
          ) : (
            grouped.map((group) => (
              <div key={group.name}>
                <p className="text-xs text-muted-foreground uppercase px-4 py-2 font-medium tracking-wide">
                  {group.name}
                </p>
                {group.items.map((item) => {
                  itemCounter++;
                  const idx = itemCounter;
                  const isActive = idx === activeIndex;
                  return (
                    <button
                      key={item.id}
                      onClick={item.action}
                      onMouseEnter={() => setActiveIndex(idx)}
                      className={`w-full px-3 py-2.5 flex items-center gap-3 cursor-pointer transition-colors mx-1 rounded-lg ${
                        isActive
                          ? "bg-primary/15 text-primary"
                          : "text-foreground hover:bg-secondary/50"
                      }`}
                      style={{ width: "calc(100% - 8px)" }}
                    >
                      <item.icon className="h-[18px] w-[18px] text-muted-foreground shrink-0" />
                      <span className="text-sm font-medium flex-1 text-left truncate">{item.label}</span>
                      {item.chips && item.chips.length > 0 && (
                        <div className="hidden sm:flex gap-1">
                          {item.chips.map((c) => (
                            <span key={c} className="text-[10px] bg-secondary text-muted-foreground px-1.5 py-0.5 rounded">{c}</span>
                          ))}
                        </div>
                      )}
                      <span className="text-xs text-muted-foreground/50 shrink-0">{item.hint}</span>
                    </button>
                  );
                })}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
