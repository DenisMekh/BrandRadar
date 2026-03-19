import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { MentionCard } from "./MentionCard";

const mockMention = {
  id: "1",
  author: "John Doe",
  source: "Web" as const,
  time: "2024-03-20T10:00:00Z",
  title: "Very interesting mention",
  text: "This is a detailed text about some brand.",
  sentiment: "positive" as const,
  sentimentScore: 0.85,
  relevant: true,
  similarCount: 2,
};

describe("MentionCard", () => {
  it("renders author and title", () => {
    render(<MentionCard mention={mockMention} />);
    expect(screen.getByText("John Doe")).toBeInTheDocument();
    expect(screen.getByText("Very interesting mention")).toBeInTheDocument();
  });

  it("renders sentiment label", () => {
    render(<MentionCard mention={mockMention} />);
    expect(screen.getByText(/Позитивная/i)).toBeInTheDocument();
  });

  it("renders similar count if present", () => {
    render(<MentionCard mention={mockMention} />);
    expect(screen.getByText("2 похожих")).toBeInTheDocument();
  });

  it("does not render similar count if 0", () => {
    render(<MentionCard mention={{ ...mockMention, similarCount: 0 }} />);
    expect(screen.queryByText(/похожих/i)).not.toBeInTheDocument();
  });

  it("toggles expanded state for long text", () => {
    const longText = "A".repeat(200);
    render(<MentionCard mention={{ ...mockMention, text: longText }} />);
    const button = screen.getByText(/показать полностью/i);
    expect(button).toBeInTheDocument();
    fireEvent.click(button);
    expect(screen.queryByText(/показать полностью/i)).not.toBeInTheDocument();
  });

  it("does not render source link if URL is missing", () => {
    render(<MentionCard mention={{ ...mockMention, url: undefined }} />);
    expect(screen.queryByText(/Открыть источник/i)).not.toBeInTheDocument();
  });
});
