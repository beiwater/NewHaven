import { RESEARCH_NODES, RESOURCES } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Lock, Play, CheckCircle2, Clock } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

const statusConfig = {
  completed: { label: "Completed", color: "bg-game-green/10 text-game-green border-game-green/20", icon: CheckCircle2 },
  in_progress: { label: "In Progress", color: "bg-game-blue/10 text-game-blue border-game-blue/20", icon: Clock },
  available: { label: "Available", color: "bg-game-yellow/10 text-game-orange border-game-orange/20", icon: Play },
  locked: { label: "Locked", color: "bg-muted text-muted-foreground border-border", icon: Lock },
};

export default function ResearchPage() {
  return (
    <div className="p-4 max-w-4xl mx-auto">
      <div className="flex items-center gap-2 mb-4">
        <h1 className="text-xl font-bold text-foreground">🔬 Research Lab</h1>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {RESEARCH_NODES.map((node, i) => {
          const cfg = statusConfig[node.status];
          const Icon = cfg.icon;
          const isLocked = node.status === "locked";

          return (
            <motion.div
              key={node.id}
              initial={{ opacity: 0, y: 15 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.06 }}
            >
              <Card className={cn(
                "p-4 border-border transition-all",
                isLocked ? "bg-muted/40 opacity-70" : "bg-card hover:shadow-md"
              )}>
                <div className="flex items-start gap-3 mb-3">
                  <div className={cn(
                    "w-11 h-11 rounded-xl flex items-center justify-center text-2xl flex-shrink-0",
                    isLocked ? "bg-muted" : "bg-accent"
                  )}>
                    {node.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className={cn("font-bold text-sm", isLocked ? "text-muted-foreground" : "text-foreground")}>{node.name}</h3>
                    <p className="text-xs text-muted-foreground mt-0.5">{node.desc}</p>
                  </div>
                </div>

                {/* Status Badge */}
                <Badge className={cn("text-[10px] mb-3", cfg.color)}>
                  <Icon className="h-3 w-3 mr-1" /> {cfg.label}
                </Badge>

                {/* Progress */}
                {node.status === "in_progress" && (
                  <div className="mb-3">
                    <div className="flex justify-between text-[10px] text-muted-foreground mb-1">
                      <span>Progress</span>
                      <span>{node.progress}%</span>
                    </div>
                    <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                      <div className="h-full bg-game-blue rounded-full transition-all" style={{ width: `${node.progress}%` }} />
                    </div>
                  </div>
                )}

                {/* Required Resources */}
                <div className="mb-3">
                  <p className="text-[10px] font-semibold text-muted-foreground mb-1">Requires</p>
                  <div className="flex flex-wrap gap-1">
                    {Object.entries(node.cost).map(([res, qty]) => (
                      <Badge key={res} variant="outline" className="text-[10px] gap-1 px-1.5 py-0">
                        {RESOURCES[res]?.icon} {qty}
                      </Badge>
                    ))}
                  </div>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-[10px] text-muted-foreground flex items-center gap-1">
                    <Clock className="h-3 w-3" /> {node.duration}
                  </span>
                  {node.status === "available" && (
                    <Button size="sm" className="bg-game-blue hover:bg-game-blue/90 text-white text-xs gap-1 rounded-full px-4">
                      <Play className="h-3 w-3" /> Start
                    </Button>
                  )}
                  {node.status === "completed" && (
                    <Badge className="bg-game-green/10 text-game-green border-game-green/20 text-[10px]">
                      <CheckCircle2 className="h-3 w-3 mr-0.5" /> Done
                    </Badge>
                  )}
                  {node.status === "in_progress" && (
                    <Badge className="bg-game-blue/10 text-game-blue border-game-blue/20 text-[10px] animate-pulse">
                      <Clock className="h-3 w-3 mr-0.5" /> Working...
                    </Badge>
                  )}
                  {isLocked && (
                    <Badge className="bg-muted text-muted-foreground text-[10px]">
                      <Lock className="h-3 w-3 mr-0.5" /> Locked
                    </Badge>
                  )}
                </div>
              </Card>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
}