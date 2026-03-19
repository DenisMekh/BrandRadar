import { useState } from "react";
import { Outlet } from "react-router-dom";
import { AppSidebar } from "@/components/AppSidebar";
import { AppHeader } from "@/components/AppHeader";
import { ConnectionStatus } from "@/components/shared/ConnectionStatus";
import { ConnectionProvider } from "@/contexts/ConnectionContext";
import { CommandPalette } from "@/components/CommandPalette";

export default function DashboardLayout() {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <ConnectionProvider>
      <div className="min-h-screen w-full bg-background">
        <AppSidebar mobileOpen={mobileOpen} onMobileClose={() => setMobileOpen(false)} />
        <div className="flex flex-col min-h-screen md:ml-16 lg:ml-60">
          <ConnectionStatus />
          <AppHeader onMenuClick={() => setMobileOpen(true)} />
          <main className="flex-1 p-4 sm:p-6">
            <div className="max-w-[1400px] mx-auto animate-page-enter">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
      <CommandPalette />
    </ConnectionProvider>
  );
}
