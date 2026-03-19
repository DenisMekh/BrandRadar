import { AlertTriangle } from "lucide-react";

export function DegradationBanner() {
  return (
    <div className="bg-warning/10 border border-warning/20 rounded-xl p-4 flex items-center gap-3">
      <AlertTriangle className="h-5 w-5 text-warning shrink-0" />
      <p className="text-sm text-warning">
        ⚠ Часть сервисов недоступна. Данные могут быть неполными.
      </p>
    </div>
  );
}
