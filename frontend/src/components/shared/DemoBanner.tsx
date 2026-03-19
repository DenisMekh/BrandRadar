import { Info } from "lucide-react";

export function DemoBanner() {
  return (
    <div className="bg-warning/10 border border-warning/20 rounded-xl px-4 py-2.5 flex items-center gap-2.5 mb-4">
      <Info className="h-4 w-4 text-warning shrink-0" />
      <p className="text-xs text-warning">
        Демо-режим: API недоступен, показаны тестовые данные
      </p>
    </div>
  );
}
