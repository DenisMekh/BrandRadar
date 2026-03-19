import { Inbox, type LucideIcon } from "lucide-react";
import { Link } from "react-router-dom";

export interface EmptyStateProps {
  icon?: LucideIcon;
  iconClassName?: string;
  title: string;
  description?: string;
  actionLabel?: string;
  actionHref?: string;
  onAction?: () => void;
}

export function EmptyState({
  icon: Icon = Inbox,
  iconClassName,
  title,
  description,
  actionLabel,
  actionHref,
  onAction,
}: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <Icon className={iconClassName || "h-12 w-12 text-muted-foreground/30 mb-4"} />
      <h3 className="text-lg font-medium text-foreground">{title}</h3>
      {description && <p className="text-sm text-muted-foreground mt-1 max-w-sm">{description}</p>}
      {actionLabel && actionHref && (
        <Link to={actionHref} className="mt-4 px-4 py-2 bg-primary text-primary-foreground text-sm rounded-lg hover:bg-primary/90 transition-colors">{actionLabel}</Link>
      )}
      {actionLabel && onAction && !actionHref && (
        <button onClick={onAction} className="mt-4 px-4 py-2 bg-primary text-primary-foreground text-sm rounded-lg hover:bg-primary/90 transition-colors cursor-pointer">{actionLabel}</button>
      )}
    </div>
  );
}
