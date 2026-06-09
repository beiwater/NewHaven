import { useState, useMemo } from "react";
import { MARKET_DATA, RESOURCES } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { TrendingUp, TrendingDown, ArrowDownLeft, ArrowUpRight, X, BarChart2, LineChart as LineIcon, Layers } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";
import { AreaChart, Area, LineChart, Line, XAxis, YAxis, ResponsiveContainer, Tooltip, ReferenceLine } from "recharts";

const TIME_RANGES = ["1h", "6h", "12h", "24h", "48h", "7d"];
const CHART_TYPES = [
  { id: "area", icon: Layers, label: "Area" },
  { id: "line", icon: LineIcon, label: "Line" },
  { id: "depth", icon: BarChart2, label: "Depth" },
];

const genChart = (base, points = 24) =>
  Array.from({ length: points }, (_, i) => ({
    t: `${i}`, price: +(base + (Math.random() - 0.48) * base * 0.18).toFixed(2),
  }));

const genDepth = (base) => {
  const steps = Array.from({ length: 12 }, (_, i) => ({
    price: +(base + (i - 6) * 0.4).toFixed(2),
    qty: Math.floor(Math.random() * 200 + 30),
  }));
  return steps;
};

const getTrend = (change) => {
  if (change > 1.5) return { label: "📈 Rising", color: "text-game-green" };
  if (change < -1.5) return { label: "📉 Falling", color: "text-game-red" };
  return { label: "➡️ Stable", color: "text-muted-foreground" };
};

function ResourceSelector({ items, selected, onSelect }) {
  return (
    <div className="space-y-0.5">
      <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-2 px-1">Resources</p>
      {items.map((item) => {
        const r = RESOURCES[item.resource];
        return (
          <button key={item.resource} onClick={() => onSelect(item.resource)}
            className={cn("w-full flex items-center gap-2 px-2 py-2 rounded-lg text-left transition-all", selected === item.resource ? "bg-accent shadow-sm" : "hover:bg-muted/50")}
          >
            <span className="text-base">{r.icon}</span>
            <div className="flex-1 min-w-0">
              <p className="font-semibold text-xs truncate">{r.name}</p>
              <p className="text-[10px] text-muted-foreground">${item.price.toFixed(1)}</p>
            </div>
            <span className={cn("text-[10px] font-bold", item.change >= 0 ? "text-game-green" : "text-game-red")}>
              {item.change >= 0 ? "+" : ""}{item.change.toFixed(1)}%
            </span>
          </button>
        );
      })}
    </div>
  );
}

function OrderBookRow({ price, qty, total, side }) {
  const pct = Math.min((total / 500) * 100, 100);
  return (
    <div className="relative flex items-center text-xs px-2 py-1">
      <div className={cn("absolute inset-y-0 right-0 opacity-10", side === "buy" ? "bg-game-green" : "bg-game-red")} style={{ width: `${pct}%` }} />
      <span className={cn("w-1/3 font-mono font-medium", side === "buy" ? "text-game-green" : "text-game-red")}>{price.toFixed(2)}</span>
      <span className="w-1/3 text-center font-mono text-muted-foreground">{qty}</span>
      <span className="w-1/3 text-right font-mono text-muted-foreground">{total}</span>
    </div>
  );
}

function SimpleMarket({ item, r }) {
  const [qty, setQty] = useState(10);
  const trend = getTrend(item.change);
  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="max-w-md mx-auto space-y-4 p-4">
      <Card className="p-5 border-border bg-card text-center">
        <span className="text-5xl block mb-2">{r.icon}</span>
        <h2 className="text-2xl font-black text-foreground">{r.name}</h2>
        <p className="text-3xl font-bold mt-1">${item.price.toFixed(2)}</p>
        <p className={cn("text-sm font-semibold mt-1", trend.color)}>{trend.label}</p>
        <div className="flex justify-center gap-4 mt-3">
          <div className="text-center">
            <p className="text-[10px] text-muted-foreground">You can buy at</p>
            <p className="text-sm font-bold text-game-green">${item.buyPrice.toFixed(2)}</p>
          </div>
          <div className="w-px bg-border" />
          <div className="text-center">
            <p className="text-[10px] text-muted-foreground">You can sell at</p>
            <p className="text-sm font-bold text-game-red">${item.sellPrice.toFixed(2)}</p>
          </div>
          <div className="w-px bg-border" />
          <div className="text-center">
            <p className="text-[10px] text-muted-foreground">Your stock</p>
            <p className="text-sm font-bold">{item.inventory}</p>
          </div>
        </div>
      </Card>

      <Card className="p-4 border-border bg-card space-y-3">
        <div>
          <label className="text-xs font-semibold text-muted-foreground">Quantity</label>
          <Input type="number" value={qty} onChange={e => setQty(Number(e.target.value))} className="h-10 text-base font-bold text-center mt-1" />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <Button className="h-12 bg-game-blue hover:bg-game-blue/90 text-white font-bold rounded-xl gap-2">
            <ArrowDownLeft className="h-4 w-4" /> Buy
            <span className="text-xs font-normal">${(item.buyPrice * qty).toFixed(0)}</span>
          </Button>
          <Button className="h-12 bg-primary hover:bg-primary/90 text-primary-foreground font-bold rounded-xl gap-2">
            <ArrowUpRight className="h-4 w-4" /> Sell
            <span className="text-xs font-normal">${(item.sellPrice * qty).toFixed(0)}</span>
          </Button>
        </div>
        <p className="text-[10px] text-center text-muted-foreground">Your cash: <span className="font-bold text-foreground">$24,580</span></p>
      </Card>
    </motion.div>
  );
}

function AdvancedMarket({ item, r }) {
  const [timeRange, setTimeRange] = useState("24h");
  const [chartType, setChartType] = useState("area");
  const [qty, setQty] = useState(10);

  const chartData = useMemo(() => {
    const pts = timeRange === "1h" ? 12 : timeRange === "6h" ? 24 : timeRange === "48h" ? 48 : timeRange === "7d" ? 52 : 36;
    return genChart(item.price, pts);
  }, [item.resource, timeRange]);

  const depthData = useMemo(() => genDepth(item.price), [item.resource]);

  const buyOrders = [
    { price: item.buyPrice - 0.5, qty: 45, total: 120 },
    { price: item.buyPrice - 1.0, qty: 80, total: 200 },
    { price: item.buyPrice - 1.5, qty: 120, total: 320 },
    { price: item.buyPrice - 2.0, qty: 200, total: 520 },
  ];
  const sellOrders = [
    { price: item.sellPrice + 0.5, qty: 30, total: 80 },
    { price: item.sellPrice + 1.0, qty: 65, total: 145 },
    { price: item.sellPrice + 1.5, qty: 95, total: 240 },
    { price: item.sellPrice + 2.0, qty: 150, total: 390 },
  ];

  const color = item.change >= 0 ? "var(--game-green)" : "var(--game-red)";

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-3">
      {/* Summary */}
      <Card className="p-3 border-border bg-card">
        <div className="flex items-center gap-2 mb-2">
          <span className="text-xl">{r.icon}</span>
          <span className="text-base font-bold">{r.name}</span>
          <span className="text-lg font-black">${item.price.toFixed(2)}</span>
          <span className={cn("text-sm font-bold flex items-center gap-0.5", item.change >= 0 ? "text-game-green" : "text-game-red")}>
            {item.change >= 0 ? <TrendingUp className="h-3.5 w-3.5" /> : <TrendingDown className="h-3.5 w-3.5" />}
            {item.change >= 0 ? "+" : ""}{item.change.toFixed(1)}%
          </span>
        </div>
        <div className="grid grid-cols-4 sm:grid-cols-8 gap-1.5">
          {[
            { label: "Best Bid", value: `$${item.buyPrice.toFixed(1)}`, color: "text-game-green" },
            { label: "Best Ask", value: `$${item.sellPrice.toFixed(1)}`, color: "text-game-red" },
            { label: "High", value: `$${item.high.toFixed(1)}` },
            { label: "Low", value: `$${item.low.toFixed(1)}` },
            { label: "Spread", value: `$${(item.buyPrice - item.sellPrice).toFixed(1)}` },
            { label: "My Stock", value: item.inventory },
            { label: "My Cash", value: "$24,580" },
            { label: "Volume", value: "4,820" },
          ].map(s => (
            <div key={s.label} className="bg-muted/50 rounded-lg p-1.5 text-center">
              <p className="text-[9px] text-muted-foreground">{s.label}</p>
              <p className={cn("text-[11px] font-bold", s.color || "text-foreground")}>{s.value}</p>
            </div>
          ))}
        </div>
      </Card>

      {/* Chart */}
      <Card className="p-3 border-border bg-card">
        <div className="flex items-center justify-between mb-2 flex-wrap gap-1.5">
          <div className="flex gap-0.5">
            {CHART_TYPES.map(ct => (
              <button key={ct.id} onClick={() => setChartType(ct.id)}
                className={cn("px-2.5 py-1 rounded-md text-xs font-medium transition-all flex items-center gap-1",
                  chartType === ct.id ? "bg-game-blue text-white" : "text-muted-foreground hover:bg-muted")}
              ><ct.icon className="h-3 w-3" />{ct.label}</button>
            ))}
          </div>
          <div className="flex gap-0.5">
            {TIME_RANGES.map(tr => (
              <button key={tr} onClick={() => setTimeRange(tr)}
                className={cn("px-2 py-1 rounded-md text-xs font-medium transition-all",
                  timeRange === tr ? "bg-accent text-foreground font-bold" : "text-muted-foreground hover:bg-muted")}
              >{tr}</button>
            ))}
          </div>
        </div>
        <div className="h-36">
          <ResponsiveContainer width="100%" height="100%">
            {chartType === "depth" ? (
              <AreaChart data={depthData}>
                <XAxis dataKey="price" tick={{ fontSize: 9 }} tickLine={false} axisLine={false} />
                <YAxis tick={{ fontSize: 9 }} width={28} tickLine={false} axisLine={false} />
                <Tooltip contentStyle={{ fontSize: 11, borderRadius: 8 }} />
                <ReferenceLine x={item.price} stroke="hsl(var(--muted-foreground))" strokeDasharray="3 3" />
                <Area type="stepAfter" dataKey="qty" stroke="hsl(var(--game-green))" fill="hsl(var(--game-green) / 0.15)" strokeWidth={1.5} />
              </AreaChart>
            ) : chartType === "line" ? (
              <LineChart data={chartData}>
                <XAxis dataKey="t" tick={{ fontSize: 9 }} tickLine={false} axisLine={false} interval="preserveStartEnd" />
                <YAxis tick={{ fontSize: 9 }} tickLine={false} axisLine={false} domain={["auto", "auto"]} width={32} />
                <Tooltip contentStyle={{ fontSize: 11, borderRadius: 8 }} />
                <Line type="monotone" dataKey="price" stroke={item.change >= 0 ? "hsl(var(--game-green))" : "hsl(var(--game-red))"} strokeWidth={2} dot={false} />
              </LineChart>
            ) : (
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={`hsl(${color})`} stopOpacity={0.2} />
                    <stop offset="95%" stopColor={`hsl(${color})`} stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="t" tick={{ fontSize: 9 }} tickLine={false} axisLine={false} interval="preserveStartEnd" />
                <YAxis tick={{ fontSize: 9 }} tickLine={false} axisLine={false} domain={["auto", "auto"]} width={32} />
                <Tooltip contentStyle={{ fontSize: 11, borderRadius: 8 }} />
                <Area type="monotone" dataKey="price" stroke={`hsl(${color})`} fill="url(#areaGrad)" strokeWidth={2} />
              </AreaChart>
            )}
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Order Books + Form + My Orders */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
        <Card className="p-3 bg-card border-border">
          <p className="text-xs font-semibold text-game-green mb-2 flex items-center gap-1"><ArrowDownLeft className="h-3 w-3" /> Buy Orders</p>
          <div className="text-[10px] text-muted-foreground flex px-2 py-1 border-b border-border/50">
            <span className="w-1/3">Price</span><span className="w-1/3 text-center">Qty</span><span className="w-1/3 text-right">Total</span>
          </div>
          {buyOrders.map((o, i) => <OrderBookRow key={i} {...o} side="buy" />)}
        </Card>
        <Card className="p-3 bg-card border-border">
          <p className="text-xs font-semibold text-game-red mb-2 flex items-center gap-1"><ArrowUpRight className="h-3 w-3" /> Sell Orders</p>
          <div className="text-[10px] text-muted-foreground flex px-2 py-1 border-b border-border/50">
            <span className="w-1/3">Price</span><span className="w-1/3 text-center">Qty</span><span className="w-1/3 text-right">Total</span>
          </div>
          {sellOrders.map((o, i) => <OrderBookRow key={i} {...o} side="sell" />)}
        </Card>
        <Card className="p-3 bg-card border-border">
          <Tabs defaultValue="buy">
            <TabsList className="w-full mb-3 bg-muted/50">
              <TabsTrigger value="buy" className="flex-1 text-xs data-[state=active]:bg-game-blue/10 data-[state=active]:text-game-blue">Limit Buy</TabsTrigger>
              <TabsTrigger value="sell" className="flex-1 text-xs data-[state=active]:bg-destructive/10 data-[state=active]:text-destructive">Limit Sell</TabsTrigger>
            </TabsList>
            {["buy", "sell"].map(side => (
              <TabsContent key={side} value={side} className="space-y-2 mt-0">
                <div><label className="text-[10px] text-muted-foreground">Price</label>
                  <Input type="number" defaultValue={side === "buy" ? item.buyPrice : item.sellPrice} className="h-8 text-sm" /></div>
                <div><label className="text-[10px] text-muted-foreground">Quantity</label>
                  <Input type="number" value={qty} onChange={e => setQty(Number(e.target.value))} className="h-8 text-sm" /></div>
                <div className="text-xs text-muted-foreground flex justify-between">
                  <span>Est. fee</span><span className="text-[10px]">~${((side === "buy" ? item.buyPrice : item.sellPrice) * qty * 0.005).toFixed(2)}</span>
                </div>
                <Button className={cn("w-full text-white text-xs", side === "buy" ? "bg-game-blue hover:bg-game-blue/90" : "bg-destructive hover:bg-destructive/90")} size="sm">
                  Place {side === "buy" ? "Buy" : "Sell"} Order
                </Button>
              </TabsContent>
            ))}
          </Tabs>
        </Card>
        <Card className="p-3 bg-card border-border">
          <p className="text-xs font-semibold text-muted-foreground mb-2">My Orders</p>
          {[
            { type: "buy", qty: 100, price: item.buyPrice - 1, filled: 65 },
            { type: "sell", qty: 50, price: item.sellPrice + 1, filled: 20 },
          ].map((o, i) => (
            <div key={i} className="mb-2 last:mb-0">
              <div className="flex items-center justify-between text-xs mb-1">
                <Badge className={cn("text-[10px] px-1.5", o.type === "buy" ? "bg-game-blue/10 text-game-blue border-game-blue/20" : "bg-destructive/10 text-destructive border-destructive/20")}>{o.type.toUpperCase()}</Badge>
                <span className="text-muted-foreground font-mono text-[10px]">{o.filled}/{o.qty}@${o.price.toFixed(1)}</span>
                <Button variant="ghost" size="icon" className="h-5 w-5 text-muted-foreground hover:text-game-red"><X className="h-3 w-3" /></Button>
              </div>
              <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
                <div className={cn("h-full rounded-full", o.type === "buy" ? "bg-game-blue" : "bg-destructive")} style={{ width: `${(o.filled / o.qty) * 100}%` }} />
              </div>
            </div>
          ))}
        </Card>
      </div>
    </motion.div>
  );
}

export default function MarketPage() {
  const [selected, setSelected] = useState("wheat");
  const [mode, setMode] = useState("simple"); // simple | advanced

  const item = MARKET_DATA.find(m => m.resource === selected);
  const r = RESOURCES[selected];

  return (
    <div className="flex flex-col lg:flex-row h-[calc(100vh-3.5rem-2.5rem)] overflow-hidden">
      {/* Mobile resource chips */}
      <div className="lg:hidden overflow-x-auto border-b border-border px-3 py-2">
        <div className="flex gap-1.5">
          {MARKET_DATA.map(m => {
            const res = RESOURCES[m.resource];
            return (
              <button key={m.resource} onClick={() => setSelected(m.resource)}
                className={cn("flex items-center gap-1 px-2.5 py-1.5 rounded-full text-xs font-medium whitespace-nowrap transition-all",
                  selected === m.resource ? "bg-accent shadow-sm" : "bg-muted/50 hover:bg-muted")}
              >{res.icon} {res.name}</button>
            );
          })}
        </div>
      </div>

      {/* Desktop sidebar */}
      <div className="hidden lg:block py-3 pl-3 w-44 flex-shrink-0 border-r border-border overflow-y-auto">
        <ResourceSelector items={MARKET_DATA} selected={selected} onSelect={setSelected} />
      </div>

      {/* Main */}
      <div className="flex-1 overflow-y-auto">
        {/* Mode toggle */}
        <div className="flex items-center justify-between px-4 py-2 border-b border-border bg-card/80 sticky top-0 z-10">
          <div className="flex items-center gap-2">
            <span className="text-lg">{r.icon}</span>
            <span className="font-bold text-sm">{r.name}</span>
            <span className={cn("text-xs font-bold", item.change >= 0 ? "text-game-green" : "text-game-red")}>
              {item.change >= 0 ? "+" : ""}{item.change.toFixed(1)}%
            </span>
          </div>
          <div className="flex items-center bg-muted rounded-full p-0.5 gap-0.5">
            <button onClick={() => setMode("simple")} className={cn("px-3 py-1 rounded-full text-xs font-semibold transition-all", mode === "simple" ? "bg-white shadow-sm text-foreground" : "text-muted-foreground")}>
              🛍️ Simple
            </button>
            <button onClick={() => setMode("advanced")} className={cn("px-3 py-1 rounded-full text-xs font-semibold transition-all", mode === "advanced" ? "bg-white shadow-sm text-foreground" : "text-muted-foreground")}>
              📊 Advanced
            </button>
          </div>
        </div>

        <AnimatePresence mode="wait">
          <motion.div key={`${selected}-${mode}`} initial={{ opacity: 0, y: 6 }} animate={{ opacity: 1, y: 0 }}>
            {mode === "simple" ? (
              <SimpleMarket item={item} r={r} />
            ) : (
              <div className="p-3">
                <AdvancedMarket item={item} r={r} />
              </div>
            )}
          </motion.div>
        </AnimatePresence>
      </div>
    </div>
  );
}