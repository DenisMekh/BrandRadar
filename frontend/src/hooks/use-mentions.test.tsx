import { renderHook, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useMentions } from "./use-mentions";
import { getMentions, type Mention } from "@/lib/api/mentions";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";

// Mock the API and config
vi.mock("@/lib/api/mentions", () => ({
  getMentions: vi.fn(),
  updateMentionStatus: vi.fn(),
}));

const { mockConfig } = vi.hoisted(() => ({
  mockConfig: { ENABLE_MOCKS: false }
}));

vi.mock("@/lib/mock-config", () => mockConfig);

vi.mock("@/lib/mock-analytics", () => ({
  getMockBrandMentions: vi.fn(() => [
    { id: "m1", title: "Test A", text: "Text A", author: "Author A", source: "Telegram", sentiment: "positive", is_relevant: true, created_at: new Date().toISOString() },
    { id: "m2", title: "Test B", text: "Text B", author: "Author B", source: "Web", sentiment: "negative", is_relevant: false, created_at: new Date().toISOString() },
  ]),
}));

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

describe("useMentions hook", () => {
  beforeEach(() => {
    queryClient.clear();
    vi.clearAllMocks();
    mockConfig.ENABLE_MOCKS = false;
  });

  it("calls getMentions API when mocks are disabled", async () => {
    const mockData = { data: [], total: 0 } as unknown as { data: unknown[]; total: number }; 
    vi.mocked(getMentions).mockResolvedValue(mockData as unknown as { data: Mention[]; total: number });

    const { result } = renderHook(() => useMentions({ limit: 10 }), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(getMentions).toHaveBeenCalledWith({ limit: 10 });
    expect(result.current.data).toEqual(mockData);
  });

  it("handles API errors", async () => {
    vi.mocked(getMentions).mockRejectedValue(new Error("API Error"));

    const { result } = renderHook(() => useMentions({}), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeDefined();
  });

  it("uses mock data when ENABLE_MOCKS is true", async () => {
    mockConfig.ENABLE_MOCKS = true;
    vi.mocked(getMentions).mockResolvedValue({ data: [], total: 0 });

    const { result } = renderHook(() => useMentions({ search: "Test A", brand_id: "b1" }), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    // Should contain filtered results
    expect(result.current.data?.total).toBe(1);
    expect(result.current.data?.data[0].title).toBe("Test A");
  });

  it("filters mock data by source and sentiment", async () => {
    mockConfig.ENABLE_MOCKS = true;
    const { result } = renderHook(() => useMentions({ source: "Web", label: "negative", brand_id: "b1" }), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.total).toBe(1);
    expect(result.current.data?.data[0].id).toBe("m2");
  });
});
