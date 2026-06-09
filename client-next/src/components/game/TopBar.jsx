import { useState } from "react";
import { PLAYER } from "@/lib/gameData";
import { Bell, Settings, LogOut } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import SettingsDrawer from "@/components/game/SettingsDrawer";
import { Link } from "react-router-dom";

export default function TopBar() {
  const p = PLAYER;
  const xpPct = (p.xp / p.xpMax) * 100;
  const [settingsOpen, setSettingsOpen] = useState(false);

  return (
    <>
      <header className="fixed top-0 left-0 right-0 z-50 h-14 bg-primary flex items-center px-3 gap-3 shadow-md">
        {/* Avatar & Name */}
        <Link to="/profile" className="flex items-center gap-2 min-w-0 hover:opacity-80 transition-opacity">
          <span className="text-2xl">{p.avatar}</span>
          <div className="hidden sm:block min-w-0">
            <p className="text-xs font-bold text-primary-foreground truncate">{p.name}</p>
            <div className="flex items-center gap-1.5">
              <span className="text-[10px] font-semibold text-primary-foreground/70">Lv.{p.level}</span>
              <div className="w-16 h-1.5 bg-primary-foreground/20 rounded-full overflow-hidden">
                <div className="h-full bg-game-yellow rounded-full transition-all" style={{ width: `${xpPct}%` }} />
              </div>
              <span className="text-[10px] text-primary-foreground/60">{p.xp}/{p.xpMax}</span>
            </div>
          </div>
        </Link>

        {/* Cash */}
        <div className="ml-auto flex items-center gap-1 bg-primary-foreground/10 rounded-full px-3 py-1">
          <span className="text-sm">💰</span>
          <span className="text-sm font-bold text-game-yellow">${p.cash.toLocaleString()}</span>
        </div>

        {/* Boosts */}
        <div className="hidden md:flex items-center gap-1">
          {p.boosts.map((b, i) => (
            <Badge key={i} className="bg-game-yellow/20 text-game-yellow border-game-yellow/30 text-[10px] px-1.5 py-0.5 animate-pulse-glow">
              {b}
            </Badge>
          ))}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="icon" className="h-8 w-8 text-primary-foreground/70 hover:text-primary-foreground hover:bg-primary-foreground/10 relative">
            <Bell className="h-4 w-4" />
            {p.notifications > 0 && (
              <span className="absolute -top-0.5 -right-0.5 w-4 h-4 bg-game-red text-[10px] text-white rounded-full flex items-center justify-center font-bold">
                {p.notifications}
              </span>
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-primary-foreground/70 hover:text-primary-foreground hover:bg-primary-foreground/10"
            onClick={() => setSettingsOpen(true)}
          >
            <Settings className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8 text-primary-foreground/70 hover:text-primary-foreground hover:bg-primary-foreground/10">
            <LogOut className="h-4 w-4" />
          </Button>
        </div>
      </header>

      <SettingsDrawer open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </>
  );
}