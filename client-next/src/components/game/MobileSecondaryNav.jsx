import { Link, useLocation } from "react-router-dom";
import { NAV_GROUPS } from "./Sidebar";
import { cn } from "@/lib/utils";

export default function MobileSecondaryNav() {
  const location = useLocation();
  const currentGroup = NAV_GROUPS.find(g => g.items.some(i => i.path === location.pathname));
  if (!currentGroup || currentGroup.items.length <= 1) return null;

  return (
    <div className="md:hidden overflow-x-auto border-b border-border bg-card/80 backdrop-blur-sm sticky top-14 z-30">
      <div className="flex gap-1 px-3 py-1.5">
        {currentGroup.items.map((item) => {
          const isActive = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                "flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium whitespace-nowrap transition-all flex-shrink-0",
                isActive
                  ? "bg-accent text-accent-foreground shadow-sm font-bold"
                  : "text-muted-foreground hover:bg-muted"
              )}
            >
              <span>{item.emoji}</span>
              <span>{item.label}</span>
            </Link>
          );
        })}
      </div>
    </div>
  );
}