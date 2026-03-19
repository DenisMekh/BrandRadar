import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react-swc";
import path from "path";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      exclude: [
        "src/components/ui/accordion.tsx",
        "src/components/ui/alert.tsx",
        "src/components/ui/aspect-ratio.tsx",
        "src/components/ui/avatar.tsx",
        "src/components/ui/badge.tsx",
        "src/components/ui/breadcrumb.tsx",
        "src/components/ui/card.tsx",
        "src/components/ui/chart.tsx",
        "src/components/ui/checkbox.tsx",
        "src/components/ui/collapsible.tsx",
        "src/components/ui/context-menu.tsx",
        "src/components/ui/dropdown-menu.tsx",
        "src/components/ui/form.tsx",
        "src/components/ui/hover-card.tsx",
        "src/components/ui/input-otp.tsx",
        "src/components/ui/menubar.tsx",
        "src/components/ui/navigation-menu.tsx",
        "src/components/ui/radio-group.tsx",
        "src/components/ui/resizable.tsx",
        "src/components/ui/scroll-area.tsx",
        "src/components/ui/select.tsx",
        "src/components/ui/table.tsx",
        "src/components/ui/tabs.tsx",
        "src/components/ui/textarea.tsx",
        "src/components/ui/progress.tsx",
        "src/components/ui/toggle-group.tsx",
        "src/main.tsx",
        "src/test/**",
        "**/*.d.ts",
      ],
    },
  },
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
});
