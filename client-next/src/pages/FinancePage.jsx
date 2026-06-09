import { FINANCE_DATA } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { BarChart, Bar, XAxis, YAxis, ResponsiveContainer, Tooltip, PieChart, Pie, Cell } from "recharts";
import { TrendingUp, TrendingDown, ArrowUpRight, ArrowDownLeft } from "lucide-react";

const f = FINANCE_DATA;

const cashflowData = [
  { name: "Mon", income: 2100, expenses: 1200 },
  { name: "Tue", income: 1800, expenses: 900 },
  { name: "Wed", income: 2400, expenses: 1500 },
  { name: "Thu", income: 1500, expenses: 800 },
  { name: "Fri", income: 2800, expenses: 1300 },
  { name: "Sat", income: 3200, expenses: 1800 },
  { name: "Sun", income: 1900, expenses: 1100 },
];

const assetBreakdown = [
  { name: "Buildings", value: 22000, fill: "hsl(var(--game-blue))" },
  { name: "Inventory", value: 12000, fill: "hsl(var(--game-green))" },
  { name: "Cash", value: 8000, fill: "hsl(var(--game-yellow))" },
];

function StatCard({ label, value, icon: Icon, trend, color }) {
  return (
    <Card className="p-4 bg-card border-border">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs text-muted-foreground font-medium">{label}</p>
          <p className={cn("text-lg font-bold mt-1", color || "text-foreground")}>${value.toLocaleString()}</p>
        </div>
        {trend && (
          <Badge className={cn("text-[10px]", trend === "up" ? "bg-game-green/10 text-game-green" : "bg-game-red/10 text-game-red")}>
            {trend === "up" ? <TrendingUp className="h-3 w-3 mr-0.5" /> : <TrendingDown className="h-3 w-3 mr-0.5" />}
            {trend === "up" ? "+12%" : "-5%"}
          </Badge>
        )}
      </div>
    </Card>
  );
}

export default function FinancePage() {
  return (
    <div className="p-4 max-w-4xl mx-auto space-y-4">
      <h1 className="text-xl font-bold text-foreground">💰 Finance</h1>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard label="Cash" value={f.cash} trend="up" color="text-game-green" />
        <StatCard label="Income" value={f.income} trend="up" />
        <StatCard label="Expenses" value={f.expenses} trend="down" />
        <StatCard label="Profit" value={f.profit} trend="up" color="text-game-blue" />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
        {/* Cashflow Chart */}
        <Card className="p-4 bg-card border-border">
          <p className="text-xs font-semibold text-muted-foreground mb-3">Weekly Cashflow</p>
          <div className="h-44">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={cashflowData}>
                <XAxis dataKey="name" tick={{ fontSize: 10 }} tickLine={false} axisLine={false} />
                <YAxis tick={{ fontSize: 10 }} tickLine={false} axisLine={false} width={35} />
                <Tooltip contentStyle={{ fontSize: 11, borderRadius: 8, border: "1px solid hsl(var(--border))" }} />
                <Bar dataKey="income" fill="hsl(var(--game-green))" radius={[4, 4, 0, 0]} barSize={14} />
                <Bar dataKey="expenses" fill="hsl(var(--game-red))" radius={[4, 4, 0, 0]} barSize={14} />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="flex justify-center gap-4 mt-2 text-[10px]">
            <span className="flex items-center gap-1"><span className="w-2 h-2 rounded bg-game-green" /> Income</span>
            <span className="flex items-center gap-1"><span className="w-2 h-2 rounded bg-game-red" /> Expenses</span>
          </div>
        </Card>

        {/* Asset Breakdown */}
        <Card className="p-4 bg-card border-border">
          <p className="text-xs font-semibold text-muted-foreground mb-3">Asset Breakdown</p>
          <div className="h-44 flex items-center justify-center">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie data={assetBreakdown} cx="50%" cy="50%" innerRadius={40} outerRadius={65} paddingAngle={3} dataKey="value">
                  {assetBreakdown.map((entry, i) => (
                    <Cell key={i} fill={entry.fill} />
                  ))}
                </Pie>
                <Tooltip contentStyle={{ fontSize: 11, borderRadius: 8, border: "1px solid hsl(var(--border))" }} />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="flex justify-center gap-4 mt-1 text-[10px]">
            {assetBreakdown.map((a) => (
              <span key={a.name} className="flex items-center gap-1">
                <span className="w-2 h-2 rounded" style={{ background: a.fill }} /> {a.name}
              </span>
            ))}
          </div>
        </Card>
      </div>

      {/* Balance Sheet & Income */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <Card className="p-4 bg-card border-border">
          <p className="text-xs font-semibold text-muted-foreground mb-3">Balance Sheet</p>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between"><span className="text-muted-foreground">Assets</span><span className="font-bold text-game-green">${f.assets.toLocaleString()}</span></div>
            <div className="flex justify-between"><span className="text-muted-foreground">Liabilities</span><span className="font-bold text-game-red">${f.liabilities.toLocaleString()}</span></div>
            <div className="border-t border-border/50 pt-2 flex justify-between"><span className="font-semibold">Equity</span><span className="font-bold">${f.equity.toLocaleString()}</span></div>
          </div>
        </Card>
        <Card className="p-4 bg-card border-border">
          <p className="text-xs font-semibold text-muted-foreground mb-3">Income Statement</p>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between"><span className="text-muted-foreground">Revenue</span><span className="font-bold text-game-green">${f.income.toLocaleString()}</span></div>
            <div className="flex justify-between"><span className="text-muted-foreground">COGS</span><span className="font-bold text-game-red">$3,200</span></div>
            <div className="flex justify-between"><span className="text-muted-foreground">Operating Exp.</span><span className="font-bold text-game-red">$2,030</span></div>
            <div className="border-t border-border/50 pt-2 flex justify-between"><span className="font-semibold">Net Profit</span><span className="font-bold text-game-green">${f.profit.toLocaleString()}</span></div>
          </div>
        </Card>
      </div>

      {/* Recent Transactions */}
      <Card className="p-4 bg-card border-border">
        <p className="text-xs font-semibold text-muted-foreground mb-3">Recent Transactions</p>
        <div className="space-y-2">
          {f.transactions.map((tx) => (
            <motion.div
              key={tx.id}
              initial={{ opacity: 0, x: -5 }}
              animate={{ opacity: 1, x: 0 }}
              className="flex items-center gap-3 px-3 py-2 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors"
            >
              <div className={cn(
                "w-7 h-7 rounded-full flex items-center justify-center flex-shrink-0",
                tx.type === "income" ? "bg-game-green/10" : "bg-game-red/10"
              )}>
                {tx.type === "income" ? <ArrowDownLeft className="h-3.5 w-3.5 text-game-green" /> : <ArrowUpRight className="h-3.5 w-3.5 text-game-red" />}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium text-foreground truncate">{tx.desc}</p>
                <p className="text-[10px] text-muted-foreground">{tx.time}</p>
              </div>
              <span className={cn("text-sm font-bold", tx.amount > 0 ? "text-game-green" : "text-game-red")}>
                {tx.amount > 0 ? "+" : ""}${Math.abs(tx.amount).toLocaleString()}
              </span>
            </motion.div>
          ))}
        </div>
      </Card>
    </div>
  );
}