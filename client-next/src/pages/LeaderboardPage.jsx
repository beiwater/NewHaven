import { LEADERBOARD } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { TrendingUp, TrendingDown, Crown } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

const rankBg = ["bg-game-yellow/20", "bg-muted/60", "bg-game-orange/15"];
const rankIcon = ["🥇", "🥈", "🥉"];

export default function LeaderboardPage() {
  return (
    <div className="p-4 max-w-2xl mx-auto">
      <div className="flex items-center gap-2 mb-4">
        <h1 className="text-xl font-bold text-foreground">🏆 Leaderboard</h1>
      </div>

      <Card className="overflow-hidden bg-card border-border">
        {LEADERBOARD.map((player, i) => (
          <motion.div
            key={player.rank}
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: i * 0.05 }}
            className={cn(
              "flex items-center gap-3 px-4 py-3 border-b border-border/50 last:border-0 transition-colors",
              player.isMe && "bg-accent/50",
              i < 3 && rankBg[i]
            )}
          >
            {/* Rank */}
            <div className="w-8 text-center flex-shrink-0">
              {i < 3 ? (
                <span className="text-lg">{rankIcon[i]}</span>
              ) : (
                <span className="text-sm font-bold text-muted-foreground">#{player.rank}</span>
              )}
            </div>

            {/* Avatar */}
            <span className="text-2xl flex-shrink-0">{player.avatar}</span>

            {/* Info */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-1.5">
                <p className={cn("text-sm font-bold truncate", player.isMe ? "text-game-blue" : "text-foreground")}>
                  {player.name}
                </p>
                {player.isMe && <Badge className="text-[8px] px-1 py-0 bg-game-blue/10 text-game-blue border-game-blue/20">You</Badge>}
              </div>
              <p className="text-xs text-muted-foreground">Level {player.level}</p>
            </div>

            {/* Net Worth */}
            <div className="text-right flex-shrink-0">
              <p className="text-sm font-bold text-foreground">${player.netWorth.toLocaleString()}</p>
              <div className="flex items-center justify-end gap-0.5">
                {player.trend === "up" ? (
                  <TrendingUp className="h-3 w-3 text-game-green" />
                ) : (
                  <TrendingDown className="h-3 w-3 text-game-red" />
                )}
                <span className={cn("text-[10px] font-medium", player.trend === "up" ? "text-game-green" : "text-game-red")}>
                  {player.trend === "up" ? "↑" : "↓"}
                </span>
              </div>
            </div>
          </motion.div>
        ))}
      </Card>
    </div>
  );
}