import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { MentionFilters } from "./MentionFilters";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";

// Mock useBrands hook
vi.mock("@/hooks/use-brands", () => ({
  useBrands: () => ({ data: [{ id: "b1", name: "Brand A" }] }),
}));

const queryClient = new QueryClient();
const wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

describe("MentionFilters", () => {
  const defaultProps = {
    search: "",
    onSearchChange: vi.fn(),
    sentiment: "all" as const,
    onSentimentChange: vi.fn(),
    onlyRelevant: false,
    onRelevantChange: vi.fn(),
    dateFrom: undefined,
    onDateFromChange: vi.fn(),
    dateTo: undefined,
    onDateToChange: vi.fn(),
    brandId: "all",
    onBrandIdChange: vi.fn(),
    source: "all",
    onSourceChange: vi.fn(),
  };

  it("calls onSearchChange when typing in search input", () => {
    render(<MentionFilters {...defaultProps} />, { wrapper });
    const input = screen.getByPlaceholderText(/Поиск по тексту/i);
    fireEvent.change(input, { target: { value: "test search" } });
    expect(defaultProps.onSearchChange).toHaveBeenCalledWith("test search");
  });

  it("calls onSentimentChange when clicking sentiment buttons", () => {
    render(<MentionFilters {...defaultProps} />, { wrapper });
    // There are two buttons (desktop/mobile), we can pick either
    const positiveButtons = screen.getAllByText("Позитивная");
    fireEvent.click(positiveButtons[0]);
    expect(defaultProps.onSentimentChange).toHaveBeenCalledWith("positive");
  });

  it("calls onRelevantChange when toggling switch", () => {
    render(<MentionFilters {...defaultProps} />, { wrapper });
    const label = screen.getByText(/Только релевантные/i);
    fireEvent.click(label);
    expect(defaultProps.onRelevantChange).toHaveBeenCalledWith(true);
  });

  it("calls onBrandIdChange when selecting a brand", () => {
    render(<MentionFilters {...defaultProps} />, { wrapper });
    const select = screen.getByDisplayValue(/Все бренды/i);
    fireEvent.change(select, { target: { value: "b1" } });
    expect(defaultProps.onBrandIdChange).toHaveBeenCalledWith("b1");
  });

  it("calls onSourceChange when selecting a source", () => {
    render(<MentionFilters {...defaultProps} />, { wrapper });
    const select = screen.getByDisplayValue(/Все источники/i);
    fireEvent.change(select, { target: { value: "Telegram" } });
    expect(defaultProps.onSourceChange).toHaveBeenCalledWith("Telegram");
  });
});
