import { useState } from "react";
import { Switch } from "@/components/ui/switch";

export function SystemSettings() {
  const [backendUrl, setBackendUrl] = useState(
    () => localStorage.getItem("backend_url") || "http://localhost:8080"
  );
  const [darkMode, setDarkMode] = useState(true);

  const handleUrlSave = () => {
    localStorage.setItem("backend_url", backendUrl);
  };

  return (
    <div className="max-w-xl space-y-6">
      <div className="bg-card border border-border rounded-xl p-5 space-y-3">
        <h4 className="text-sm font-medium text-foreground">URL бэкенда</h4>
        <div className="flex gap-2">
          <input
            value={backendUrl}
            onChange={(e) => setBackendUrl(e.target.value)}
            className="flex-1 bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
          />
          <button onClick={handleUrlSave} className="px-4 py-2 bg-primary text-primary-foreground text-sm rounded-lg hover:bg-primary/90 transition-colors cursor-pointer">Сохранить</button>
        </div>
      </div>

      <div className="bg-card border border-border rounded-xl p-5 flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium text-foreground">Тёмная тема</h4>
          <p className="text-xs text-muted-foreground mt-0.5">Включена по умолчанию</p>
        </div>
        <Switch checked={darkMode} onCheckedChange={setDarkMode} />
      </div>

      <div className="bg-card border border-border rounded-xl p-5 space-y-2">
        <h4 className="text-sm font-medium text-foreground">Информация</h4>
        <div className="grid grid-cols-2 gap-y-1.5 text-sm">
          <span className="text-muted-foreground">Версия</span>
          <span className="text-foreground">v1.0.0</span>
        </div>
      </div>
    </div>
  );
}
