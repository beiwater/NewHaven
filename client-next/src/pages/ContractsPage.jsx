import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Clock, CheckCircle2, AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { RESOURCES } from "@/lib/gameData";

const CONTRACTS = [
  { id: 1, client: "Harbor Inn 🏨", items: [{ resource: "bread", qty: 30 }, { resource: "butter", qty: 10 }], reward: 1500, deadline: "2h left", status: "active", progress: 60 },
  { id: 2, client: "Town Festival 🎪", items: [{ resource: "cake", qty: 15 }, { resource: "cookie", qty: 50 }], reward: 3200, deadline: "5h left", status: "active", progress: 25 },
  { id: 3, client: "Ship Captain ⚓", items: [{ resource: "fish", qty: 100 }], reward: 2000, deadline: "Completed", status: "completed", progress: 100 },
  { id: 4, client: "Village School 🏫", items: [{ resource: "milk", qty: 20 }, { resource: "apple", qty: 30 }], reward: 800, deadline: "1d left", status: "active", progress: 10 },
  { id: 5, client: "Royal Palace 👑", items: [{ resource: "cake", qty: 50 }, { resource: "coffee", qty: 30 }, { resource: "honey", qty: 20 }], reward: 8500, deadline: "Locked", status: "locked", progress: 0 },
];

const statusCfg = {
  active: { color: "bg-game-blue/10 text-game-blue border-game-blue/20", icon: Clock },
  completed: { color: "bg-game-green/10 text-game-green border-game-green/20", icon: CheckCircle2 },
  locked: { color: "bg-muted text-muted-foreground border-border", icon: AlertCircle },
};

export default function ContractsPage() {
  return (
    <div className="p-4 max-w-3xl mx-auto">
      <div className="flex items-center gap-2 mb-4">
        <h1 className="text-xl font-bold text-foreground">📜 Contracts</h1>
      </div>

      <div className="space-y-3">
        {CONTRACTS.map((contract, i) => {
          const cfg = statusCfg[contract.status];
          const Icon = cfg.icon;
          const isLocked = contract.status === "locked";

          return (
            <motion.div
              key={contract.id}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.06 }}
            >
              <Card className={cn(
                "p-4 border-border transition-all",
                isLocked ? "bg-muted/30 opacity-70" : "bg-card hover:shadow-md"
              )}>
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="font-bold text-sm text-foreground">{contract.client}</h3>
                    <Badge className={cn("text-[10px] mt-1", cfg.color)}>
                      <Icon className="h-3 w-3 mr-0.5" /> {contract.deadline}
                    </Badge>
                  </div>
                  <div className="text-right">
                    <p className="text-xs text-muted-foreground">Reward</p>
                    <p className="text-sm font-bold text-game-green">💰 ${contract.reward.toLocaleString()}</p>
                  </div>
                </div>

                {/* Required Items */}
                <div className="flex flex-wrap gap-1.5 mb-3">
                  {contract.items.map((item, j) => (
                    <Badge key={j} variant="outline" className="text-xs gap-1">
                      {RESOURCES[item.resource]?.icon} {item.qty} {RESOURCES[item.resource]?.name}
                    </Badge>
                  ))}
                </div>

                {/* Progress */}
                {contract.status === "active" && (
                  <div>
                    <div className="flex justify-between text-[10px] text-muted-foreground mb-1">
                      <span>Progress</span>
                      <span>{contract.progress}%</span>
                    </div>
                    <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                      <div className="h-full bg-game-blue rounded-full transition-all" style={{ width: `${contract.progress}%` }} />
                    </div>
                  </div>
                )}

                {contract.status === "active" && (
                  <Button size="sm" className="mt-3 bg-game-blue hover:bg-game-blue/90 text-white text-xs gap-1 rounded-full px-4">
                    Deliver Items
                  </Button>
                )}
              </Card>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
}