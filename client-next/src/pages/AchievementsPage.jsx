import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

const ACHIEVEMENTS = [
  { id: 1, icon: "🌾", name: "First Harvest", desc: "Harvest your first crop", unlocked: true },
  { id: 2, icon: "🍞", name: "Baker's Dozen", desc: "Bake 13 loaves of bread", unlocked: true },
  { id: 3, icon: "💰", name: "Big Spender", desc: "Earn $10,000 total", unlocked: true },
  { id: 4, icon: "🏗️", name: "Master Builder", desc: "Build 5 structures", unlocked: true },
  { id: 5, icon: "⚓", name: "Trade Captain", desc: "Complete 50 trades", unlocked: false, progress: 32, total: 50 },
  { id: 6, icon: "🎂", name: "Pastry Chef", desc: "Bake 100 cakes", unlocked: false, progress: 45, total: 100 },
  { id: 7, icon: "🐟", name: "Master Angler", desc: "Catch 200 fish", unlocked: false, progress: 120, total: 200 },
  { id: 8, icon: "🏆", name: "Top 3", desc: "Reach top 3 on leaderboard", unlocked: false, progress: 0, total: 1 },
  { id: 9, icon: "👑", name: "Tycoon", desc: "Reach $100,000 net worth", unlocked: false, progress: 42000, total: 100000 },
  { id: 10, icon: "🌟", name: "Perfectionist", desc: "Get 100% quality on all items", unlocked: false, progress: 4, total: 10 },
];

export default function AchievementsPage() {
  const unlocked = ACHIEVEMENTS.filter((a) => a.unlocked).length;

  return (
    <div className="p-4 max-w-3xl mx-auto">
      <div className="flex items-center gap-2 mb-1">
        <h1 className="text-xl font-bold text-foreground">⭐ Achievements</h1>
      </div>
      <p className="text-xs text-muted-foreground mb-4">{unlocked}/{ACHIEVEMENTS.length} unlocked</p>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {ACHIEVEMENTS.map((ach, i) => {
          const pct = ach.unlocked ? 100 : ach.total ? (ach.progress / ach.total) * 100 : 0;
          return (
            <motion.div
              key={ach.id}
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: i * 0.04 }}
            >
              <Card className={cn(
                "p-4 border-border transition-all",
                ach.unlocked ? "bg-card" : "bg-muted/30 opacity-75"
              )}>
                <div className="flex items-center gap-3">
                  <div className={cn(
                    "w-11 h-11 rounded-xl flex items-center justify-center text-2xl flex-shrink-0",
                    ach.unlocked ? "bg-game-yellow/20" : "bg-muted"
                  )}>
                    {ach.icon}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-1.5">
                      <h3 className="font-bold text-sm text-foreground truncate">{ach.name}</h3>
                      {ach.unlocked && <Badge className="bg-game-green/10 text-game-green border-game-green/20 text-[8px] px-1 py-0">✓</Badge>}
                    </div>
                    <p className="text-xs text-muted-foreground">{ach.desc}</p>
                    {!ach.unlocked && ach.total && (
                      <div className="mt-1.5">
                        <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
                          <div className="h-full bg-game-yellow rounded-full transition-all" style={{ width: `${pct}%` }} />
                        </div>
                        <p className="text-[10px] text-muted-foreground mt-0.5">{ach.progress}/{ach.total}</p>
                      </div>
                    )}
                  </div>
                </div>
              </Card>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
}