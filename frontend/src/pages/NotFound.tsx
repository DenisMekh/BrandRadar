import { Link } from "react-router-dom";
import { ShieldOff } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";

const NotFound = () => {
  usePageTitle("404");

  return (
    <div className="flex min-h-screen items-center justify-center bg-background animate-page-enter">
      <div className="text-center space-y-4">
        <ShieldOff className="h-16 w-16 text-muted-foreground/30 mx-auto" />
        <h1 className="text-5xl font-bold text-foreground">404</h1>
        <p className="text-lg text-muted-foreground">Страница не найдена</p>
        <Link
          to="/"
          className="inline-block mt-2 px-6 py-2.5 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:bg-primary/90 transition-colors cursor-pointer focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          На главную
        </Link>
      </div>
    </div>
  );
};

export default NotFound;
