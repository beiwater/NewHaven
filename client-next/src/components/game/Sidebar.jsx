import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";

export const NAV_GROUPS = [
  {
    id: "town",
    label: "Town",
    emoji: "🏘️",
    items: [
      { path: "/", label: "Map", emoji: "🗺️" },
      { path: "/build", label: "Build", emoji: "🔨" },
      { path: "/collection", label: "Collect", emoji: "🧺" },
      { path: "/warehouse", label: "Warehouse", emoji: "📦" },
    ],
  },
  {
    id: "trade",
    label: "Trade",
    emoji: "📈",
    items: [
      { path: "/market", label: "Market", emoji: "📊" },
      { path: "/orders", label: "My Orders", emoji: "📋" },
      { path: "/contracts", label: "Contracts", emoji: "📜" },
      { path: "/finance", label: "Finance", emoji: "💰" },
    ],
  },
  {
    id: "growth",
    label: "Growth",
    emoji: "🌱",
    items: [
      { path: "/research", label: "Research", emoji: "🔬" },
      { path: "/executives", label: "Executives", emoji: "👔" },
      { path: "/achievements", label: "Achievements", emoji: "⭐" },
      { path: "/leaderboard", label: "Leaderboard", emoji: "🏆" },
    ],
  },
  {
    id: "social",
    label: "Social",
    emoji: "💬",
    items: [
      { path: "/chat", label: "Chat", emoji: "🌐" },
      { path: "/messages", label: "Messages", emoji: "📨" },
    ],
  },
  {
    id: "help",
    label: "Help",
    emoji: "📚",
    items: [
      { path: "/wiki", label: "Wiki", emoji: "📚" },
      { path: "/settings", label: "Settings", emoji: "⚙️" },
    ],
  },
];

function getActiveGroup(pathname) {
  return NAV_GROUPS.find(g => g.items.some(i => i.path === pathname))?.id || "town";
}

export default function Sidebar() {
  const location = useLocation();
  const [activeGroup, setActiveGroup] = useState(() => getActiveGroup(location.pathname));

  const currentGroup = NAV_GROUPS.find(g => g.id === activeGroup);

  return (
    <>
      {/* ── Desktop: icon rail + submenu ── */}
      <aside className="hidden md:flex fixed top-14 left-0 bottom-10 z-40">
        {/* Primary icon rail */}
        <div className="w-14 bg-card border-r border-border flex flex-col py-2 gap-0.5 items-center overflow-y-auto">
          {NAV_GROUPS.map((g) => {
            const isActive = g.id === activeGroup;
            const hasCurrentPage = g.items.some(i => i.path === location.pathname);
            return (
              <button
                key={g.id}
                onClick={() => setActiveGroup(g.id)}
                className={cn(
                  "w-10 h-10 rounded-xl flex flex-col items-center justify-center gap-0.5 transition-all text-[10px] font-medium",
                  isActive
                    ? "bg-accent text-accent-foreground shadow-sm"
                    : hasCurrentPage
                    ? "bg-accent/50 text-accent-foreground"
                    : "text-muted-foreground hover:bg-muted"
                )}
              >
                <span className="text-lg leading-none">{g.emoji}</span>
              </button>
            );
          })}
        </div>

        {/* Secondary submenu */}
        <AnimatePresence mode="wait">
          <motion.div
            key={activeGroup}
            initial={{ opacity: 0, x: -8 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -8 }}
            transition={{ duration: 0.15 }}
            className="w-36 bg-card/95 border-r border-border flex flex-col py-3 px-2 gap-0.5"
          >
            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider px-2 mb-1">
              {currentGroup?.label}
            </p>
            {currentGroup?.items.map((item) => {
              const isActive = location.pathname === item.path;
              return (
                <Link
                  key={item.path}
                  to={item.path}
                  className={cn(
                    "flex items-center gap-2 px-2.5 py-2 rounded-lg text-sm font-medium transition-all",
                    isActive
                      ? "bg-accent text-accent-foreground shadow-sm"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground"
                  )}
                >
                  <span className="text-base flex-shrink-0">{item.emoji}</span>
                  <span className="truncate text-xs">{item.label}</span>
                </Link>
              );
            })}
          </motion.div>
        </AnimatePresence>
      </aside>

      {/* ── Mobile: bottom primary nav ── */}
      <nav className="md:hidden fixed bottom-0 left-0 right-0 z-50 bg-card border-t border-border flex items-center justify-around px-2 py-1">
        {NAV_GROUPS.map((g) => {
          const isActive = g.id === activeGroup || g.items.some(i => i.path === location.pathname);
          const hasCurrentPage = g.items.some(i => i.path === location.pathname);
          return (
            <Link
              key={g.id}
              to={g.items[0].path}
              onClick={() => setActiveGroup(g.id)}
              className={cn(
                "flex flex-col items-center gap-0.5 px-2 py-1.5 rounded-xl text-[10px] font-medium transition-all",
                hasCurrentPage ? "text-primary" : "text-muted-foreground"
              )}
            >
              <span className="text-xl">{g.emoji}</span>
              <span>{g.label}</span>
            </Link>
          );
        })}
      </nav>

      {/* ── Mobile: secondary horizontal tabs (rendered via context) ── */}
    </>
  );
}