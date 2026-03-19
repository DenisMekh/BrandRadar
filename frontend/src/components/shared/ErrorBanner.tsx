import { useState, useEffect, useCallback } from "react";
import { AlertCircle, RefreshCw } from "lucide-react";

interface ErrorBannerProps {
  message?: string;
  onRetry?: () => void;
  autoRetrySeconds?: number;
}

export function ErrorBanner({ message = "Ошибка загрузки данных", onRetry, autoRetrySeconds = 10 }: ErrorBannerProps) {
  const [countdown, setCountdown] = useState(autoRetrySeconds);

  const doRetry = useCallback(() => {
    onRetry?.();
    setCountdown(autoRetrySeconds);
  }, [onRetry, autoRetrySeconds]);

  useEffect(() => {
    if (!onRetry) return;
    const id = setInterval(() => {
      setCountdown((c) => {
        if (c <= 1) {
          doRetry();
          return autoRetrySeconds;
        }
        return c - 1;
      });
    }, 1000);
    return () => clearInterval(id);
  }, [doRetry, autoRetrySeconds, onRetry]);

  return (
    <div className="bg-destructive/10 border border-destructive/20 rounded-xl p-4 flex items-center gap-3">
      <AlertCircle className="h-5 w-5 text-destructive shrink-0" />
      <p className="text-sm text-destructive flex-1">{message}</p>
      {onRetry && (
        <div className="flex items-center gap-3 shrink-0">
          <span className="text-xs text-destructive/60">
            Повтор через {countdown} сек...
          </span>
          <button
            onClick={doRetry}
            className="text-sm text-destructive font-medium hover:text-destructive/80 transition-colors shrink-0 flex items-center gap-1 cursor-pointer"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            Повторить
          </button>
        </div>
      )}
    </div>
  );
}
