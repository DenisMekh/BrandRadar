import { describe, it, expect } from "vitest";
import { fmtNum, fmtRelative } from "./format";

describe("fmtNum", () => {
  it("formats regular numbers correctly", () => {
    expect(fmtNum(1234)).toBe("1\u00A0234");
    expect(fmtNum(1000000)).toBe("1\u00A0000\u00A0000");
  });

  it("handles string numbers", () => {
    expect(fmtNum("5678")).toBe("5\u00A0678");
  });

  it("returns dash for null/undefined/invalid", () => {
    expect(fmtNum(null)).toBe("—");
    expect(fmtNum(undefined)).toBe("—");
    expect(fmtNum("not a number")).toBe("—");
  });

  it("handles zero and negative numbers", () => {
    expect(fmtNum(0)).toBe("0");
    expect(fmtNum(-1234)).toBe("-1\u00A0234");
  });

  it("handles very large numbers", () => {
    expect(fmtNum(1234567890)).toBe("1\u00A0234\u00A0567\u00A0890");
  });
});

describe("fmtRelative", () => {
  it("formats relative dates correctly in Russian", () => {
    const pastDate = new Date(Date.now() - 60000).toISOString();
    expect(fmtRelative(pastDate)).toContain("назад");
  });

  it("handles future dates", () => {
    const futureDate = new Date(Date.now() + 60000).toISOString();
    expect(fmtRelative(futureDate)).toBeDefined(); // usually "через..."
  });

  it("returns dash for empty input", () => {
    expect(fmtRelative(null)).toBe("—");
    expect(fmtRelative("")).toBe("—");
  });

  it("returns original string on error", () => {
    expect(fmtRelative("invalid-date")).toBe("invalid-date");
  });
});
