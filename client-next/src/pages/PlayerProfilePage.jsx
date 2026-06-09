import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { TrendingUp, MessageCircle, UserPlus, Settings, Package } from "lucide-react";
import { PLAYER, BUILDINGS, FINANCE_DATA, WAREHOUSE_ITEMS, EXECUTIVES, RESEARCH_NODES } from "@/lib/gameData";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { Link } from "react-router-dom";

const MY_BUILDINGS = [
  { id: "farm", name: "Farm", level: 3, region: "New Harbor", status: "producing", icon: "🌾" },
  { id: "barn", name: "Barn", level: 2, region: "New Harbor", status: "producing", icon: "🐄" },
  { id: "mill", name: "Mill", level: 1, region: "New Harbor", status: "idle", icon: "⚙️" },
  { id: "bakery", name: "Bakery", level: 2, region: "New Harbor", status: "ready", icon: "🍞" },
  { id: "cafe", name: "Café", level: 1, region: "New Harbor", status: "selling", icon: "☕" },
  { id: "warehouse", name: "Warehouse", level: 1, region: "New Harbor", status: "storing", icon: "📦" },
];

const STATUS_STYLE = {
  producing: "bg-game-blue/10 text-game-blue border-game-blue/20",
  idle: "bg-muted text-muted-foreground border-border",
  ready: "bg-game-green/10 text-game-green border-game-green/20",
  selling: "bg-game-orange/10 text-game-orange border-game-orange/20",
  storing: "bg-purple-100 text-purple-600 border-purple-200",
};

function StatCard({ label, value, sub, icon, color }) {
  return (
    <div className="bg-muted/40 rounded-xl p-3 text-center">
      {icon && <div className="text-xl mb-1">{icon}</div>}
      <p className={cn("text-lg font-black", color || "text-foreground")}>{value}</p>
      <p className="text-[10px] text-muted-foreground font-medium">{label}</p>
      {sub && <p className="text-[10px] text-muted-foreground">{sub}</p>}
    </div>
  );
}

export default function PlayerProfilePage() {
  const p = PLAYER;
  const xpPct = (p.xp / p.xpMax) * 100;
  const warehouseUsed = WAREHOUSE_ITEMS.reduce((a, i) => a + i.quantity, 0);
  const warehouseTotal = WAREHOUSE_ITEMS.reduce((a, i) => a + i.capacity, 0);
  const warehousePct = (warehouseUsed / warehouseTotal) * 100;
  const completedResearch = RESEARCH_NODES.filter(r => r.status === "completed").length;

  return (
    <div className="p-4 max-w-2xl mx-auto space-y-4">
      {/* Profile Header */}
      <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }}>
        <Card className="p-5 border-border bg-card overflow-hidden relative">
          <div className="absolute inset-0 bg-gradient-to-br from-primary/5 to-accent/30 pointer-events-none" />
          <div className="relative flex items-start gap-4">
            <div className="relative flex-shrink-0">
              <div className="w-16 h-16 rounded-2xl bg-accent flex items-center justify-center text-4xl shadow-md">{p.avatar}</div>
              <span className="absolute -bottom-1 -right-1 w-4 h-4 bg-game-green rounded-full border-2 border-card" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-start justify-between">
                <div>
                  <h1 className="text-lg font-black text-foreground">{p.name}</h1>
                  <p className="text-sm text-muted-foreground">Mochi Foods Co. · Harbor Tycoon</p>
                  <div className="flex items-center gap-2 mt-1 flex-wrap">
                    <Badge className="bg-game-green/10 text-game-green border-game-green/20 text-xs">🟢 Online</Badge>
                    <Badge variant="outline" className="text-xs">Level {p.level}</Badge>
                    <Badge variant="outline" className="text-xs">📍 New Harbor</Badge>
                  </div>
                </div>
                <Link to="/settings">
                  <Button variant="outline" size="icon" className="h-8 w-8 flex-shrink-0">
                    <Settings className="h-3.5 w-3.5" />
                  </Button>
                </Link>
              </div>
              <div className="mt-3">
                <div className="flex justify-between text-xs mb-1">
                  <span className="text-muted-foreground">XP Progress</span>
                  <span className="font-semibold">{p.xp}/{p.xpMax}</span>
                </div>
                <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                  <div className="h-full bg-game-yellow rounded-full transition-all" style={{ width: `${xpPct}%` }} />
                </div>
              </div>
            </div>
          </div>
        </Card>
      </motion.div>

      {/* Company Overview */}
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }}>
        <Card className="p-4 border-border bg-card">
          <p className="text-xs font-bold text-muted-foreground mb-3">🏢 Company Overview</p>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
            <StatCard label="Cash" value={`$${p.cash.toLocaleString()}`} color="text-game-green" icon="💰" />
            <StatCard label="Net Worth" value="$42,000" color="text-foreground" icon="📊" />
            <StatCard label="Buildings" value={MY_BUILDINGS.length} icon="🏗️" />
            <StatCard label="Research" value={`${completedResearch}/${RESEARCH_NODES.length}`} icon="🔬" />
          </div>
        </Card>
      </motion.div>

      {/* Economy Stats */}
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }}>
        <Card className="p-4 border-border bg-card">
          <p className="text-xs font-bold text-muted-foreground mb-3">📈 Economy Stats</p>
          <div className="grid grid-cols-2 gap-3">
            {[
              { label: "Est. Hourly Income", value: "$140/hr", color: "text-game-green", icon: "⏱" },
              { label: "Hourly Costs", value: "$62/hr", color: "text-game-red", icon: "💸" },
              { label: "Best Seller", value: "🎂 Cake", icon: "🏆" },
              { label: "Most Produced", value: "🌾 Wheat", icon: "🏭" },
            ].map(s => (
              <div key={s.label} className="flex items-center gap-2 bg-muted/40 rounded-xl p-3">
                <span className="text-lg">{s.icon}</span>
                <div>
                  <p className="text-xs text-muted-foreground">{s.label}</p>
                  <p className={cn("text-sm font-bold", s.color || "text-foreground")}>{s.value}</p>
                </div>
              </div>
            ))}
          </div>
          <div className="mt-3 flex items-center gap-2 bg-muted/40 rounded-xl p-3">
            <TrendingUp className="h-4 w-4 text-game-green flex-shrink-0" />
            <div>
              <p className="text-xs text-muted-foreground">Market Reputation</p>
              <p className="text-sm font-bold text-foreground">⭐⭐⭐ Trusted Merchant</p>
            </div>
          </div>
        </Card>
      </motion.div>

      {/* Buildings Summary */}
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15 }}>
        <Card className="p-4 border-border bg-card">
          <p className="text-xs font-bold text-muted-foreground mb-3">🏗️ My Buildings</p>
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
            {MY_BUILDINGS.map(b => (
              <div key={b.id} className="bg-muted/40 rounded-xl p-2.5 flex items-center gap-2">
                <span className="text-xl">{b.icon}</span>
                <div className="min-w-0 flex-1">
                  <p className="text-xs font-semibold text-foreground truncate">{b.name}</p>
                  <div className="flex items-center gap-1 mt-0.5">
                    <Badge className="text-[9px] px-1 h-3.5 bg-white/80 text-foreground border-border/50">Lv.{b.level}</Badge>
                    <Badge className={cn("text-[9px] px-1 h-3.5", STATUS_STYLE[b.status])}>{b.status}</Badge>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </motion.div>

      {/* Inventory Summary */}
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }}>
        <Card className="p-4 border-border bg-card">
          <p className="text-xs font-bold text-muted-foreground mb-2">📦 Inventory</p>
          <div className="flex justify-between text-xs mb-1">
            <span>Warehouse Usage</span>
            <span className={warehousePct > 80 ? "text-game-red font-bold" : "text-muted-foreground"}>{warehouseUsed}/{warehouseTotal}</span>
          </div>
          <div className="w-full h-2.5 bg-muted rounded-full overflow-hidden mb-3">
            <div className={cn("h-full rounded-full transition-all", warehousePct > 80 ? "bg-game-red" : warehousePct > 60 ? "bg-game-yellow" : "bg-game-green")} style={{ width: `${warehousePct}%` }} />
          </div>
          <div className="flex flex-wrap gap-1.5">
            {WAREHOUSE_ITEMS.slice(0, 6).map(item => {
              const RESOURCE_ICONS = { wheat: "🌾", corn: "🌽", milk: "🥛", egg: "🥚", flour: "🫘", bread: "🍞", fish: "🐟", honey: "🍯", coffee: "☕", cookie: "🍪" };
              return (
                <span key={item.resource} className="bg-muted/50 text-xs px-2 py-1 rounded-full border border-border">
                  {RESOURCE_ICONS[item.resource]} {item.quantity}
                </span>
              );
            })}
          </div>
        </Card>
      </motion.div>

      {/* Recent Achievements */}
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.25 }}>
        <Card className="p-4 border-border bg-card">
          <p className="text-xs font-bold text-muted-foreground mb-3">⭐ Recent Achievements</p>
          <div className="space-y-2">
            {[
              { icon: "🌾", name: "First Harvest", desc: "Harvested your first crop" },
              { icon: "🍞", name: "Baker's Dozen", desc: "Baked 13 loaves of bread" },
              { icon: "💰", name: "Big Spender", desc: "Earned $10,000 total" },
            ].map(a => (
              <div key={a.name} className="flex items-center gap-3 bg-game-yellow/5 border border-game-yellow/20 rounded-xl p-2.5">
                <span className="text-2xl">{a.icon}</span>
                <div>
                  <p className="text-xs font-bold text-foreground">{a.name}</p>
                  <p className="text-[10px] text-muted-foreground">{a.desc}</p>
                </div>
                <Badge className="ml-auto bg-game-green/10 text-game-green border-game-green/20 text-[9px]">✓</Badge>
              </div>
            ))}
          </div>
        </Card>
      </motion.div>

      {/* Social Actions */}
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3 }}>
        <Card className="p-4 border-border bg-card">
          <p className="text-xs font-bold text-muted-foreground mb-3">Social</p>
          <div className="grid grid-cols-2 gap-2">
            <Link to="/chat">
              <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2 text-xs rounded-xl" size="sm">
                <MessageCircle className="h-3.5 w-3.5" /> Send Message
              </Button>
            </Link>
            <Button variant="outline" className="gap-2 text-xs rounded-xl" size="sm">
              <UserPlus className="h-3.5 w-3.5" /> Add Contact
            </Button>
          </div>
        </Card>
      </motion.div>
    </div>
  );
}