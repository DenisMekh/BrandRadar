/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useMemo } from "react";
import { useHealth } from "@/hooks/use-health";
import { mockHealth } from "@/lib/mock-data";
import type { HealthStatus } from "@/lib/api/health";

type ConnectionState = "online" | "degraded" | "offline";

interface ConnectionContextValue {
  state: ConnectionState;
  health: HealthStatus | undefined;
  isOffline: boolean;
  failedDeps: string[];
}

const Ctx = createContext<ConnectionContextValue>({
  state: "online",
  health: undefined,
  isOffline: false,
  failedDeps: [],
});

export function ConnectionProvider({ children }: { children: React.ReactNode }) {
  const { data, isError } = useHealth();

  const value = useMemo<ConnectionContextValue>(() => {
    // API unreachable — data fell back to mockHealth
    if (isError || data === mockHealth) {
      return { state: "offline", health: data, isOffline: true, failedDeps: [] };
    }
    if (!data) return { state: "online", health: undefined, isOffline: false, failedDeps: [] };

    // dependencies is now { [key: string]: string } e.g. { "redis": "ok", "postgres": "degraded" }
    const depsMap = data.dependencies ?? {};
    const failedDeps = Object.entries(depsMap)
      .filter(([_, status]) => status !== "ok")
      .map(([name]) => name);

    const state: ConnectionState = data.status !== "ok" ? "degraded" : "online";
    return { state, health: data, isOffline: false, failedDeps };
  }, [data, isError]);

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export const useConnection = () => useContext(Ctx);
