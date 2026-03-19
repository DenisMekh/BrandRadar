import { useConnection } from "@/contexts/ConnectionContext";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

interface OfflineGuardProps {
  children: React.ReactElement;
  className?: string;
}

/**
 * Wraps a button/action. When offline, disables it and shows a tooltip.
 */
export function OfflineGuard({ children, className }: OfflineGuardProps) {
  const { isOffline } = useConnection();

  if (!isOffline) return children;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className={cn("inline-flex", className)}>
          <span className="pointer-events-none opacity-40">{children}</span>
        </span>
      </TooltipTrigger>
      <TooltipContent>Недоступно в офлайн-режиме</TooltipContent>
    </Tooltip>
  );
}
