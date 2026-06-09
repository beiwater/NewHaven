import { MARKET_DATA, RESOURCES } from "@/lib/gameData";
import { Link } from "react-router-dom";
import { TrendingUp, TrendingDown } from "lucide-react";

export default function PriceTicker() {
  const items = [...MARKET_DATA, ...MARKET_DATA]; // duplicate for seamless scroll

  return (
    <div className="hidden md:block fixed bottom-0 left-16 lg:left-44 right-0 h-10 bg-card border-t border-border z-40 overflow-hidden">
      <div className="flex items-center h-full animate-ticker whitespace-nowrap">
        {items.map((item, i) => {
          const r = RESOURCES[item.resource];
          const isUp = item.change >= 0;
          return (
            <Link
              key={`${item.resource}-${i}`}
              to="/market"
              className="inline-flex items-center gap-1.5 px-4 h-full hover:bg-muted/50 transition-colors cursor-pointer border-r border-border/50"
            >
              <span className="text-sm">{r.icon}</span>
              <span className="text-xs font-semibold text-foreground">{r.name}</span>
              <span className="text-xs font-bold text-foreground">${item.price.toFixed(1)}</span>
              <span className={`text-[10px] font-bold flex items-center gap-0.5 ${isUp ? "text-game-green" : "text-game-red"}`}>
                {isUp ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                {isUp ? "+" : ""}{item.change.toFixed(1)}%
              </span>
            </Link>
          );
        })}
      </div>
    </div>
  );
}