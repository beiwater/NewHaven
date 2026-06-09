import { EXECUTIVES } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { GraduationCap, TrendingUp, Percent, BadgeDollarSign } from "lucide-react";
import { motion } from "framer-motion";

export default function ExecutivesPage() {
  return (
    <div className="p-4 max-w-4xl mx-auto">
      <div className="flex items-center gap-2 mb-4">
        <h1 className="text-xl font-bold text-foreground">👔 Executives</h1>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {EXECUTIVES.map((exec, i) => {
          const levelPct = (exec.level / exec.maxLevel) * 100;
          return (
            <motion.div
              key={exec.id}
              initial={{ opacity: 0, y: 15 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.08 }}
            >
              <Card className="p-4 bg-card border-border hover:shadow-md transition-shadow">
                {/* Header */}
                <div className="flex items-center gap-3 mb-4">
                  <div className="w-14 h-14 rounded-2xl bg-accent flex items-center justify-center text-3xl">
                    {exec.avatar}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="font-bold text-foreground">{exec.name}</h3>
                    <p className="text-xs text-muted-foreground">{exec.role}</p>
                    <Badge variant="outline" className="text-[10px] mt-1 gap-1">
                      <BadgeDollarSign className="h-3 w-3" /> ${exec.salary}/day
                    </Badge>
                  </div>
                </div>

                {/* Stats */}
                <div className="grid grid-cols-3 gap-2 mb-4">
                  <div className="bg-muted/40 rounded-lg p-2 text-center">
                    <p className="text-[10px] text-muted-foreground">Production</p>
                    <p className="text-sm font-bold text-game-green">+{exec.prodBonus}%</p>
                  </div>
                  <div className="bg-muted/40 rounded-lg p-2 text-center">
                    <p className="text-[10px] text-muted-foreground">Sales</p>
                    <p className="text-sm font-bold text-game-blue">+{exec.salesBonus}%</p>
                  </div>
                  <div className="bg-muted/40 rounded-lg p-2 text-center">
                    <p className="text-[10px] text-muted-foreground">Discount</p>
                    <p className="text-sm font-bold text-game-purple">-{exec.mgmtDiscount}%</p>
                  </div>
                </div>

                {/* Level Progress */}
                <div className="mb-3">
                  <div className="flex items-center justify-between text-xs mb-1">
                    <span className="text-muted-foreground">Level {exec.level}</span>
                    <span className="text-muted-foreground">{exec.level}/{exec.maxLevel}</span>
                  </div>
                  <div className="w-full h-2.5 bg-muted rounded-full overflow-hidden">
                    <div className="h-full bg-game-yellow rounded-full transition-all" style={{ width: `${levelPct}%` }} />
                  </div>
                </div>

                {/* Train Button */}
                <Button className="w-full bg-primary hover:bg-primary/90 text-primary-foreground text-xs gap-1.5 rounded-full" size="sm">
                  <GraduationCap className="h-3.5 w-3.5" /> Train — $200
                </Button>
              </Card>
            </motion.div>
          );
        })}

        {/* Recruit New */}
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.35 }}
        >
          <Card className="p-4 bg-muted/20 border-border border-dashed flex flex-col items-center justify-center min-h-[260px] hover:bg-muted/30 transition-colors cursor-pointer">
            <span className="text-4xl mb-2">➕</span>
            <p className="text-sm font-bold text-foreground">Recruit Executive</p>
            <p className="text-xs text-muted-foreground mt-1">Hire new talent for your team</p>
          </Card>
        </motion.div>
      </div>
    </div>
  );
}