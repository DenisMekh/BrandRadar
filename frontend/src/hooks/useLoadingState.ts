import { useState, useEffect, useCallback } from "react";

export function useLoadingState(delay = 1500) {
  const [isLoading, setIsLoading] = useState(true);
  const [hasError, setHasError] = useState(false);

  const load = useCallback(() => {
    setIsLoading(true);
    setHasError(false);
    const timer = setTimeout(() => setIsLoading(false), delay);
    return () => clearTimeout(timer);
  }, [delay]);

  useEffect(() => {
    const cleanup = load();
    return cleanup;
  }, [load]);

  const retry = () => {
    load();
  };

  const simulateError = () => {
    setIsLoading(false);
    setHasError(true);
  };

  return { isLoading, hasError, retry, simulateError };
}
