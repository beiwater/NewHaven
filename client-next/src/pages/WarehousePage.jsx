import { WAREHOUSE_ITEMS, RESOURCES } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

export default function WarehousePage() {
  const totalUsed = WAREHOUSE_ITEMS.reduce((s, i) => s + i.quantity, 0);
  const totalCap = WAREHOUSE_ITEMS.reduce((s, i) => s + i.capacity, 0);
  const totalPct = (totalUsed / totalCap) * 100;

  return (
    <div className="p-4 max-w-4xl mx-auto">
      <div className="flex items-center gap-2 mb-4">
        <h1 className="text-xl font-bold text-foreground">📦 Warehouse</h1>
      </div>

      {/* Total Capacity */}
      <Card className="p-4 bg-card border-border mb-4">
        <div className="flex items-center justify-between mb-2">
          <p className="text-xs font-semibold text-muted-foreground">Total Capacity</p>
          <p className="text-sm font-bold text-foreground">{totalUsed.toLocaleString()} / {totalCap.toLocaleString()}</p>
        </div>
        <div className="w-full h-3 bg-muted rounded-full overflow-hidden">
          <div
            className={cn(
              "h-full rounded-full transition-all",
              totalPct > 90 ? "bg-game-red" : totalPct > 70 ? "bg-game-yellow" : "bg-game-green"
            )}
            style={{ width: `${totalPct}%` }}
          />
        </div>
      </Card>

      {/* Resource Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {WAREHOUSE_ITEMS.map((item, i) => {
          const r = RESOURCES[item.resource];
          const pct = (item.quantity / item.capacity) * 100;
          return (
            <motion.div
              key={item.resource}
              initial={{ opacity: 0, y: 15 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.04 }}
            >
              <Card className="p-4 bg-card border-border hover:shadow-md transition-shadow">
                <div className="flex items-center gap-3 mb-3">
                  <div className="w-10 h-10 rounded-xl bg-accent flex items-center justify-center text-xl">
                    {r.icon}
                  </div>
                  <div className="flex-1">
                    <h3 className="font-bold text-sm text-foreground">{r.name}</h3>
                    <p className="text-xs text-muted-foreground">{item.quantity} / {item.capacity}</p>
                  </div>
                  <Badge className={cn(
                    "text-[10px]",
                    item.quality >= 95 ? "bg-game-green/10 text-game-green border-game-green/20" :
                    item.quality >= 85 ? "bg-game-blue/10 text-game-blue border-game-blue/20" :
                    "bg-game-yellow/10 text-game-orange border-game-orange/20"
                  )}>
                    ⭐ {item.quality}%
                  </Badge>
                </div>
                <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                  <div
                    className={cn(
                      "h-full rounded-full transition-all",
                      pct > 90 ? "bg-game-red" : pct > 70 ? "bg-game-yellow" : "bg-game-green"
                    )}
                    style={{ width: `${pct}%` }}
                  />
                </div>
              </Card>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
}