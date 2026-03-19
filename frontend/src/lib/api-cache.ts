const PREFIX = "cache:";

export function clearCache() {
  const keys = Object.keys(localStorage);
  keys.forEach(key => {
    if (key.startsWith(PREFIX)) {
      try { localStorage.removeItem(key); } catch { /* ignore */ }
    }
  });
}

export function setCache(key: string, data: unknown) {
  try {
    localStorage.setItem(PREFIX + key, JSON.stringify({ data, ts: Date.now() }));
  } catch {
    // quota exceeded — try to clear all cache to make room for new one or vital UI state
    clearCache();
    try {
      localStorage.setItem(PREFIX + key, JSON.stringify({ data, ts: Date.now() }));
    } catch {
      // Still failing? Just ignore.
    }
  }
}

export function getCache<T = unknown>(key: string): { data: T; ts: number } | null {
  try {
    const raw = localStorage.getItem(PREFIX + key);
    if (!raw) return null;
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function fmtCacheTime(ts: number): string {
  const d = new Date(ts);
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}
