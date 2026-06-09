import { useState } from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Slider } from "@/components/ui/slider";
import { X, ArrowUp, Move, Trash2, Play, Zap, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { RESOURCES, BUILDINGS } from "@/lib/gameData";

// --- Farm ---
function FarmDetail({ building, plot }) {
  const [crop, setCrop] = useState("wheat");
  const crops = [
    { id: "wheat", icon: "🌾", name: "Wheat", time: "30m", yield: 100 },
    { id: "corn", icon: "🌽", name: "Corn", time: "25m", yield: 80 },
    { id: "vegetable", icon: "🥦", name: "Vegetables", time: "45m", yield: 60 },
  ];
  return (
    <div className="space-y-3">
      <p className="text-xs font-semibold text-muted-foreground">☀️ Weather Bonus: <span className="text-game-yellow font-bold">+15% Spring Boost</span></p>
      <div>
        <p className="text-xs font-semibold text-muted-foreground mb-1.5">Choose Crop</p>
        <div className="grid grid-cols-3 gap-1.5">
          {crops.map(c => (
            <button key={c.id} onClick={() => setCrop(c.id)} className={cn("rounded-xl p-2 text-center border transition-all text-xs", crop === c.id ? "bg-accent border-primary" : "bg-muted/50 border-border")}>
              <div className="text-xl mb-0.5">{c.icon}</div>
              <div className="font-semibold">{c.name}</div>
              <div className="text-[10px] text-muted-foreground">⏱ {c.time} · ×{c.yield}</div>
            </button>
          ))}
        </div>
      </div>
      <Button className="w-full bg-game-green hover:bg-game-green/90 text-white gap-2" size="sm">
        <Play className="h-3.5 w-3.5" /> Start Growing
      </Button>
    </div>
  );
}

// --- Barn ---
function BarnDetail({ building, plot }) {
  const [careMeter] = useState(78);
  return (
    <div className="space-y-3">
      <div className="bg-muted/50 rounded-xl p-3">
        <div className="flex justify-between text-xs mb-1"><span className="font-semibold">Animal Care</span><span className={cn(careMeter > 60 ? "text-game-green" : "text-game-red")}>{careMeter}%</span></div>
        <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
          <div className="h-full bg-game-green rounded-full" style={{ width: `${careMeter}%` }} />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-2 text-xs">
        <div className="bg-muted/50 rounded-lg p-2 text-center"><p className="text-muted-foreground">Feed Cost</p><p className="font-bold">$120/cycle</p></div>
        <div className="bg-muted/50 rounded-lg p-2 text-center"><p className="text-muted-foreground">Output</p><p className="font-bold">🥛50 🥚80</p></div>
      </div>
      <p className="text-[10px] text-muted-foreground">🐄 Stable output with slightly higher operating cost. Keep animal care above 70% for best results.</p>
      <Button className="w-full bg-game-green hover:bg-game-green/90 text-white gap-2" size="sm">
        <Play className="h-3.5 w-3.5" /> Feed & Produce
      </Button>
    </div>
  );
}

// --- Mill ---
function MillDetail() {
  const [inputQty, setInputQty] = useState([50]);
  const outputQty = Math.floor(inputQty[0] * 0.8);
  const efficiency = 80;
  return (
    <div className="space-y-3">
      <div>
        <div className="flex justify-between text-xs mb-1"><span className="font-semibold text-muted-foreground">🌾 Input Wheat</span><span>{inputQty[0]} units</span></div>
        <Slider value={inputQty} onValueChange={setInputQty} min={10} max={200} step={10} className="mt-2" />
      </div>
      <div className="bg-muted/50 rounded-xl p-3 flex items-center justify-between text-xs">
        <div className="text-center"><p className="text-muted-foreground">Input</p><p className="font-bold text-lg">🌾 {inputQty[0]}</p></div>
        <span className="text-muted-foreground">→</span>
        <div className="text-center"><p className="text-muted-foreground">Output</p><p className="font-bold text-lg">🫘 {outputQty}</p></div>
      </div>
      <div><div className="flex justify-between text-[10px] text-muted-foreground mb-1"><span>Processing Efficiency</span><span>{efficiency}%</span></div>
        <div className="w-full h-1.5 bg-muted rounded-full"><div className="h-full bg-game-blue rounded-full" style={{ width: `${efficiency}%` }} /></div>
      </div>
      <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm"><Play className="h-3.5 w-3.5" /> Process Grain</Button>
    </div>
  );
}

// --- Kitchen ---
function KitchenDetail() {
  const [recipe, setRecipe] = useState("dough");
  const [batch, setBatch] = useState([1]);
  const recipes = [
    { id: "dough", icon: "🫓", name: "Dough", inputs: ["🫘 Flour x30", "🥛 Milk x10"], time: "20m" },
    { id: "butter", icon: "🧈", name: "Butter", inputs: ["🥛 Milk x20"], time: "15m" },
    { id: "cheese", icon: "🧀", name: "Cheese", inputs: ["🥛 Milk x40", "🥚 Egg x10"], time: "40m" },
  ];
  const r = recipes.find(r => r.id === recipe);
  return (
    <div className="space-y-3">
      <div>
        <p className="text-xs font-semibold text-muted-foreground mb-1.5">Recipe</p>
        <div className="space-y-1">
          {recipes.map(rec => (
            <button key={rec.id} onClick={() => setRecipe(rec.id)} className={cn("w-full flex items-center gap-2.5 p-2 rounded-lg border text-left transition-all text-xs", recipe === rec.id ? "bg-accent border-primary" : "bg-muted/30 border-border")}>
              <span className="text-lg">{rec.icon}</span>
              <div><p className="font-semibold">{rec.name}</p><p className="text-[10px] text-muted-foreground">{rec.inputs.join(" + ")}</p></div>
              <span className="ml-auto text-[10px] text-muted-foreground">⏱ {rec.time}</span>
            </button>
          ))}
        </div>
      </div>
      <div><div className="flex justify-between text-xs mb-1"><span className="font-semibold text-muted-foreground">Batch Size</span><span>×{batch[0]}</span></div>
        <Slider value={batch} onValueChange={setBatch} min={1} max={5} step={1} /></div>
      <Button className="w-full bg-game-orange hover:bg-game-orange/90 text-white gap-2" size="sm"><Play className="h-3.5 w-3.5" /> Start Cooking</Button>
    </div>
  );
}

// --- Bakery ---
function BakeryDetail() {
  const ingredients = [
    { name: "Flour", icon: "🫘", qty: 30, available: true },
    { name: "Eggs", icon: "🥚", qty: 8, available: true },
    { name: "Butter", icon: "🧈", qty: 10, available: false },
    { name: "Sugar", icon: "🍬", qty: 15, available: true },
  ];
  const allAvailable = ingredients.every(i => i.available);
  return (
    <div className="space-y-3">
      <p className="text-xs font-semibold text-muted-foreground">Ingredient Checklist — 🎂 Cake</p>
      <div className="space-y-1.5">
        {ingredients.map((ing) => (
          <div key={ing.name} className={cn("flex items-center gap-2 px-3 py-2 rounded-lg text-xs", ing.available ? "bg-game-green/5 border border-game-green/20" : "bg-game-red/5 border border-game-red/20")}>
            <span className="text-base">{ing.icon}</span>
            <span className="flex-1 font-medium">{ing.name}</span>
            <span className="text-muted-foreground">×{ing.qty}</span>
            <span>{ing.available ? "✅" : "❌"}</span>
          </div>
        ))}
      </div>
      {!allAvailable && <p className="text-[10px] text-game-red">⚠️ Missing ingredients! Buy Butter from the market.</p>}
      <div className="bg-muted/50 rounded-lg p-2 text-xs flex justify-between">
        <span className="text-muted-foreground">Quality Bonus</span><span className="text-game-purple font-bold">+12% (Chef Bruno)</span>
      </div>
      <Button className={cn("w-full gap-2", allAvailable ? "bg-game-blue hover:bg-game-blue/90 text-white" : "bg-muted text-muted-foreground cursor-not-allowed")} size="sm" disabled={!allAvailable}>
        <Play className="h-3.5 w-3.5" /> Bake Cake
      </Button>
    </div>
  );
}

// --- Café ---
function CafeDetail() {
  const [coffeeMenu, setCoffeeMenu] = useState(true);
  const [cakeMenu, setCakeMenu] = useState(true);
  const happiness = coffeeMenu && cakeMenu ? 95 : coffeeMenu || cakeMenu ? 70 : 30;
  return (
    <div className="space-y-3">
      <p className="text-xs font-semibold text-muted-foreground">Menu Builder</p>
      {[{ label: "☕ Coffee", state: coffeeMenu, set: setCoffeeMenu, note: "Base revenue" }, { label: "🎂 Cake", state: cakeMenu, set: setCakeMenu, note: "Combo bonus" }].map(m => (
        <div key={m.label} className={cn("flex items-center justify-between p-3 rounded-xl border cursor-pointer", m.state ? "bg-accent border-primary" : "bg-muted/30 border-border")} onClick={() => m.set(!m.state)}>
          <div><p className="font-semibold text-sm">{m.label}</p><p className="text-[10px] text-muted-foreground">{m.note}</p></div>
          <span>{m.state ? "✅" : "⬜"}</span>
        </div>
      ))}
      {coffeeMenu && cakeMenu && <Badge className="bg-game-yellow/10 text-game-yellow border-game-yellow/20 text-xs">✨ Combo Bonus: +25% Revenue</Badge>}
      <div><div className="flex justify-between text-xs mb-1"><span className="text-muted-foreground">Customer Happiness</span><span className={happiness >= 90 ? "text-game-green font-bold" : "text-game-yellow"}>{happiness}%</span></div>
        <div className="w-full h-2 bg-muted rounded-full"><div className={cn("h-full rounded-full", happiness >= 90 ? "bg-game-green" : happiness >= 60 ? "bg-game-yellow" : "bg-game-red")} style={{ width: `${happiness}%` }} /></div>
      </div>
      <Button className="w-full bg-game-orange hover:bg-game-orange/90 text-white gap-2" size="sm"><Play className="h-3.5 w-3.5" /> Open Café</Button>
    </div>
  );
}

// --- Food Truck ---
function FoodTruckDetail() {
  const [location, setLocation] = useState("harbor");
  const locations = [
    { id: "harbor", name: "Harbor Front", bonus: "+20%", saturation: "Low", fuel: "$30" },
    { id: "festival", name: "Food Festival", bonus: "+60%", saturation: "High", fuel: "$60" },
    { id: "market", name: "Central Market", bonus: "+10%", saturation: "Medium", fuel: "$20" },
  ];
  const loc = locations.find(l => l.id === location);
  return (
    <div className="space-y-3">
      <p className="text-xs font-semibold text-muted-foreground">Select Location</p>
      <div className="space-y-1.5">
        {locations.map(l => (
          <button key={l.id} onClick={() => setLocation(l.id)} className={cn("w-full flex items-center justify-between p-2.5 rounded-xl border text-xs transition-all", location === l.id ? "bg-accent border-primary" : "bg-muted/30 border-border")}>
            <span className="font-medium">{l.name}</span>
            <div className="flex items-center gap-2">
              <Badge className="text-[10px] bg-game-green/10 text-game-green border-game-green/20">{l.bonus}</Badge>
              <span className="text-muted-foreground">⛽ {l.fuel}</span>
            </div>
          </button>
        ))}
      </div>
      {loc.saturation === "High" && <p className="text-[10px] text-game-yellow flex items-center gap-1"><Zap className="h-3 w-3" /> High saturation — competition may reduce revenue.</p>}
      <div className="grid grid-cols-2 gap-2 text-xs">
        <div className="bg-muted/50 rounded-lg p-2 text-center"><p className="text-muted-foreground">Revenue Bonus</p><p className="font-bold text-game-green">{loc.bonus}</p></div>
        <div className="bg-muted/50 rounded-lg p-2 text-center"><p className="text-muted-foreground">Fuel Cost</p><p className="font-bold">{loc.fuel}</p></div>
      </div>
      <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm"><Play className="h-3.5 w-3.5" /> Drive & Sell</Button>
    </div>
  );
}

// --- Restaurant ---
function RestaurantDetail() {
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-2 text-xs">
        {[{ label: "Brand Rating", value: "⭐⭐⭐⭐", color: "text-game-yellow" }, { label: "Menu Quality", value: "88%", color: "text-game-green" }, { label: "Exec Bonus", value: "+15%", color: "text-game-purple" }, { label: "Daily Revenue", value: "$1,200", color: "text-foreground" }].map(s => (
          <div key={s.label} className="bg-muted/50 rounded-lg p-2 text-center"><p className="text-muted-foreground">{s.label}</p><p className={cn("font-bold", s.color)}>{s.value}</p></div>
        ))}
      </div>
      <p className="text-[10px] text-muted-foreground">Set your menu items to maximize quality score. Higher scores attract premium customers.</p>
      <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm"><Play className="h-3.5 w-3.5" /> Open for Dinner</Button>
    </div>
  );
}

// --- Trading Hub ---
function TradingHubDetail() {
  const [autoTrade, setAutoTrade] = useState(false);
  const [spread, setSpread] = useState([2]);
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between p-3 rounded-xl border bg-muted/30">
        <div><p className="font-semibold text-sm">Auto-Trading</p><p className="text-[10px] text-muted-foreground">Bot buys low and sells high automatically</p></div>
        <button onClick={() => setAutoTrade(!autoTrade)} className={cn("w-10 h-5 rounded-full transition-all", autoTrade ? "bg-game-green" : "bg-muted")}>
          <div className={cn("w-4 h-4 bg-white rounded-full transition-all m-0.5", autoTrade ? "translate-x-5" : "")} />
        </button>
      </div>
      <div><div className="flex justify-between text-xs mb-1"><span className="font-semibold text-muted-foreground">Bid-Ask Spread</span><span>{spread[0]}%</span></div>
        <Slider value={spread} onValueChange={setSpread} min={1} max={10} step={0.5} />
        <p className="text-[10px] text-muted-foreground mt-1">Higher spread = more profit per trade but fewer fills.</p>
      </div>
      {autoTrade && <Badge className="bg-game-yellow/10 text-game-yellow border-game-yellow/20 text-xs">⚠️ Auto-trading uses market liquidity. Risk of loss in volatile markets.</Badge>}
      <Button className={cn("w-full gap-2", autoTrade ? "bg-game-green hover:bg-game-green/90 text-white" : "bg-game-blue hover:bg-game-blue/90 text-white")} size="sm">
        {autoTrade ? "✅ Bot Running" : "Start Trading Bot"}
      </Button>
    </div>
  );
}

// --- Market Stall ---
function MarketStallDetail() {
  const [price, setPrice] = useState(35);
  return (
    <div className="space-y-3">
      <div>
        <p className="text-xs font-semibold text-muted-foreground mb-1.5">Product to Sell</p>
        <div className="grid grid-cols-3 gap-1.5">
          {["🍞 Bread", "🍪 Cookies", "🧀 Cheese"].map(p => (
            <button key={p} className="bg-accent rounded-lg p-2 text-xs font-medium text-center">{p}</button>
          ))}
        </div>
      </div>
      <div className="flex items-center gap-3">
        <span className="text-xs text-muted-foreground">Price: $</span>
        <input type="number" value={price} onChange={e => setPrice(Number(e.target.value))} className="flex-1 h-8 text-sm border border-border rounded-lg px-2 bg-muted/30" />
      </div>
      <div className="grid grid-cols-2 gap-2 text-xs">
        <div className="bg-muted/50 rounded-lg p-2 text-center"><p className="text-muted-foreground">Customer Flow</p><p className="font-bold text-game-green">Stable 🟢</p></div>
        <div className="bg-muted/50 rounded-lg p-2 text-center"><p className="text-muted-foreground">Risk Level</p><p className="font-bold text-game-blue">Low</p></div>
      </div>
      <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm"><Play className="h-3.5 w-3.5" /> Open Stall</Button>
    </div>
  );
}

// --- Warehouse ---
function WarehouseDetail() {
  const used = 1955; const total = 2500;
  const pct = (used / total) * 100;
  return (
    <div className="space-y-3">
      <div>
        <div className="flex justify-between text-xs mb-1"><span className="font-semibold">Capacity</span><span className={pct > 80 ? "text-game-red font-bold" : "text-muted-foreground"}>{used}/{total}</span></div>
        <div className="w-full h-3 bg-muted rounded-full overflow-hidden">
          <div className={cn("h-full rounded-full transition-all", pct > 80 ? "bg-game-red" : pct > 60 ? "bg-game-yellow" : "bg-game-green")} style={{ width: `${pct}%` }} />
        </div>
      </div>
      {pct > 80 && <Badge className="bg-game-red/10 text-game-red border-game-red/20 text-xs">⚠️ Almost full — upgrade or sell resources!</Badge>}
      <div className="grid grid-cols-3 gap-1.5 text-xs">
        {[{ cat: "Raw", pct: 45, color: "bg-game-green" }, { cat: "Processed", pct: 30, color: "bg-game-blue" }, { cat: "Finished", pct: 25, color: "bg-game-purple" }].map(c => (
          <div key={c.cat} className="bg-muted/50 rounded-lg p-2 text-center">
            <div className="w-full h-1.5 rounded-full bg-muted mb-1"><div className={cn("h-full rounded-full", c.color)} style={{ width: `${c.pct}%` }} /></div>
            <p className="text-muted-foreground">{c.cat}</p>
          </div>
        ))}
      </div>
      <Button className="w-full bg-game-yellow hover:bg-game-yellow/90 text-foreground gap-2" size="sm"><ArrowUp className="h-3.5 w-3.5" /> Upgrade Storage (+200)</Button>
    </div>
  );
}

// --- Shop ---
function ShopDetail() {
  return (
    <div className="space-y-3">
      <p className="text-xs font-semibold text-muted-foreground">Product Shelf</p>
      {[{ icon: "🎂", name: "Cake", price: 72, demand: "High", trend: "📈" }, { icon: "☕", name: "Coffee", price: 60, demand: "Medium", trend: "→" }, { icon: "🍪", name: "Cookies", price: 28, demand: "High", trend: "📈" }].map(p => (
        <div key={p.name} className="flex items-center gap-3 p-2.5 rounded-xl bg-muted/40 border border-border text-xs">
          <span className="text-xl">{p.icon}</span>
          <div className="flex-1"><p className="font-semibold">{p.name}</p><p className="text-muted-foreground">Demand: {p.demand} {p.trend}</p></div>
          <span className="font-bold">${p.price}</span>
        </div>
      ))}
      <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm"><Play className="h-3.5 w-3.5" /> Open Shop</Button>
    </div>
  );
}

// --- Generic fallback ---
function GenericDetail({ building, plot }) {
  return (
    <div className="space-y-2">
      {building.produces.length > 0 && (
        <div>
          <p className="text-xs font-semibold text-muted-foreground mb-1.5">Produces</p>
          <div className="flex flex-wrap gap-1.5">
            {building.produces.map(r => <Badge key={r} variant="outline" className="gap-1 text-xs">{RESOURCES[r]?.icon} {RESOURCES[r]?.name}</Badge>)}
          </div>
        </div>
      )}
      <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm"><Play className="h-3.5 w-3.5" /> Start Production</Button>
    </div>
  );
}

const MECHANIC_MAP = {
  farm: FarmDetail,
  barn: BarnDetail,
  mill: MillDetail,
  kitchen: KitchenDetail,
  bakery: BakeryDetail,
  cafe: CafeDetail,
  food_truck: FoodTruckDetail,
  restaurant: RestaurantDetail,
  trading_hub: TradingHubDetail,
  market_stall: MarketStallDetail,
  warehouse: WarehouseDetail,
  shop: ShopDetail,
};

export default function BuildingDetailPanel({ plot, onClose }) {
  const building = plot?.building ? BUILDINGS.find(b => b.id === plot.building) : null;
  if (!building) return null;

  const MechanicComponent = MECHANIC_MAP[building.id] || GenericDetail;

  return (
    <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: 20 }}>
      <Card className="p-4 border-border bg-card">
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center gap-2">
            <span className="text-3xl">{building.icon}</span>
            <div>
              <h3 className="font-bold text-foreground">{building.name}</h3>
              <p className="text-xs text-muted-foreground">{building.desc}</p>
            </div>
          </div>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="flex items-center gap-2 mb-3">
          <Badge className="bg-game-green/10 text-game-green border-game-green/20">Level {plot.level}</Badge>
          <Badge variant="outline" className="text-muted-foreground">Plot #{plot.id}</Badge>
        </div>

        {/* Unique mechanic */}
        <div className="border-t border-border pt-3 mb-3">
          <MechanicComponent building={building} plot={plot} />
        </div>

        {/* Common actions */}
        <div className="grid grid-cols-2 gap-2 border-t border-border pt-3">
          <Button variant="outline" className="gap-1.5 text-xs" size="sm">
            <ArrowUp className="h-3 w-3" /> Upgrade
          </Button>
          <Button variant="outline" className="gap-1.5 text-xs" size="sm">
            <Move className="h-3 w-3" /> Move
          </Button>
        </div>
        <Button variant="outline" className="w-full gap-1.5 text-xs text-game-red border-game-red/30 hover:bg-game-red/10 mt-2" size="sm">
          <Trash2 className="h-3 w-3" /> Demolish
        </Button>
      </Card>
    </motion.div>
  );
}