import { renderHook, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useAlertConfigs, useAlertHistory } from "./use-alerts";
import { getAllAlertConfigs, getAlerts, type AlertConfig } from "@/lib/api/alerts";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";

vi.mock("@/lib/api/alerts", () => ({
  getAllAlertConfigs: vi.fn(),
  updateAlertConfig: vi.fn(),
  getAlerts: vi.fn(),
}));

const { mockConfig } = vi.hoisted(() => ({
  mockConfig: { ENABLE_MOCKS: false }
}));
vi.mock("@/lib/mock-config", () => mockConfig);

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

describe("useAlerts hooks", () => {
  beforeEach(() => {
    queryClient.clear();
    vi.clearAllMocks();
    mockConfig.ENABLE_MOCKS = false;
  });

  it("calls getAllAlertConfigs API when mocks are disabled", async () => {
    const mockData = [{ id: "a1", name: "Alert 1" }] as unknown as AlertConfig[];
    vi.mocked(getAllAlertConfigs).mockResolvedValue(mockData);

    const { result } = renderHook(() => useAlertConfigs(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(getAllAlertConfigs).toHaveBeenCalled();
    expect(result.current.data).toEqual(mockData);
  });

  it("uses mock data in useAlertHistory when ENABLE_MOCKS is true", async () => {
    mockConfig.ENABLE_MOCKS = true;
    const { result } = renderHook(() => useAlertHistory("b1"), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeDefined();
    expect(Array.isArray(result.current.data)).toBe(true);
  });
});
