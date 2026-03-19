import { renderHook, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useBrands } from "./use-brands";
import { getBrands, type Brand } from "@/lib/api/brands";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";

vi.mock("@/lib/api/brands", () => ({
  getBrands: vi.fn(),
  createBrand: vi.fn(),
  updateBrand: vi.fn(),
  deleteBrand: vi.fn(),
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

describe("useBrands hook", () => {
  beforeEach(() => {
    queryClient.clear();
    vi.clearAllMocks();
    mockConfig.ENABLE_MOCKS = false;
  });

  it("calls getBrands API when mocks are disabled", async () => {
    const mockBrandsData = [{ id: "b1", name: "Brand 1" }] as unknown as Brand[];
    vi.mocked(getBrands).mockResolvedValue(mockBrandsData);

    const { result } = renderHook(() => useBrands("p1"), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(getBrands).toHaveBeenCalledWith("p1");
    expect(result.current.data).toEqual(mockBrandsData);
  });

  it("uses mock data when ENABLE_MOCKS is true", async () => {
    mockConfig.ENABLE_MOCKS = true;
    const { result } = renderHook(() => useBrands(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeDefined();
    expect(Array.isArray(result.current.data)).toBe(true);
    expect(getBrands).not.toHaveBeenCalled();
  });
});
