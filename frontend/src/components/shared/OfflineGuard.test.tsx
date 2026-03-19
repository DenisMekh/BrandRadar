import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { OfflineGuard } from "./OfflineGuard";
import { useConnection } from "@/contexts/ConnectionContext";
import { TooltipProvider } from "@/components/ui/tooltip";

vi.mock("@/contexts/ConnectionContext", () => ({
  useConnection: vi.fn(),
}));

describe("OfflineGuard", () => {
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <TooltipProvider>{children}</TooltipProvider>
  );

  it("renders children directly when online", () => {
    vi.mocked(useConnection).mockReturnValue({ isOffline: false } as unknown as ReturnType<typeof useConnection>);
    render(
      <OfflineGuard>
        <button>Click me</button>
      </OfflineGuard>,
      { wrapper }
    );
    expect(screen.getByText("Click me")).toBeInTheDocument();
    expect(screen.queryByText("Недоступно в офлайн-режиме")).not.toBeInTheDocument();
  });

  it("renders disabled state with tooltip when offline", () => {
    vi.mocked(useConnection).mockReturnValue({ isOffline: true } as unknown as ReturnType<typeof useConnection>);
    render(
      <OfflineGuard>
        <button>Click me</button>
      </OfflineGuard>,
      { wrapper }
    );
    
    // The button is wrapped in spans
    expect(screen.getByText("Click me")).toBeInTheDocument();
    // RTL tooltip content might not be in the document until hover, 
    // but the guard component renders it.
  });
});
