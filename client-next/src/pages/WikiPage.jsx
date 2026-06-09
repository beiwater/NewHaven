import { useState, useMemo } from "react";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, ChevronRight, ArrowLeft, BookOpen, ExternalLink } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";
import { Link } from "react-router-dom";

const WIKI_RESOURCES = [
  { id: "grain", icon: "🌾", name: "Grain", category: "Raw", tier: 1, basePrice: 12.5, producedBy: ["Farm"], usedIn: ["Flour","Bread"], tradable: true, tip: "The foundation of your food chain. Always keep a steady supply." },
  { id: "milk", icon: "🥛", name: "Dairy Milk", category: "Raw", tier: 1, basePrice: 15.0, producedBy: ["Barn"], usedIn: ["Butter","Cheese"], tradable: true, tip: "Barn production is stable but has a slightly higher cost per cycle." },
  { id: "flour", icon: "🫘", name: "Flour", category: "Processed", tier: 2, basePrice: 22.0, producedBy: ["Mill"], usedIn: ["Bread","Cake","Dough"], tradable: true, tip: "Flour is the most traded processed good. Keep your Mill running." },
  { id: "dough", icon: "🫓", name: "Dough", category: "Processed", tier: 2, basePrice: 18.0, producedBy: ["Kitchen"], usedIn: ["Bread","Pizza"], tradable: false, tip: "Used internally. Not tradable but essential for bread chains." },
  { id: "butter", icon: "🧈", name: "Butter", category: "Processed", tier: 2, basePrice: 20.0, producedBy: ["Kitchen"], usedIn: ["Cake","Pastry"], tradable: true, tip: "Butter prices spike during festival events. Stock up in advance." },
  { id: "sugar", icon: "🍬", name: "Sugar", category: "Raw", tier: 1, basePrice: 10.0, producedBy: ["Farm"], usedIn: ["Cake","Cookie"], tradable: true, tip: "Sugar from Farm. Essential for high-value desserts." },
  { id: "cheese", icon: "🧀", name: "Cheese", category: "Processed", tier: 2, basePrice: 28.0, producedBy: ["Kitchen"], usedIn: ["Pizza","Restaurant"], tradable: true, tip: "Cheese has high margins. A Kitchen focused on Cheese can be very profitable." },
  { id: "steak", icon: "🥩", name: "Steak", category: "Raw", tier: 2, basePrice: 45.0, producedBy: ["Barn"], usedIn: ["Restaurant"], tradable: true, tip: "High value but only useful for Restaurants. Premium income." },
  { id: "cake", icon: "🎂", name: "Cake", category: "Finished", tier: 3, basePrice: 65.0, producedBy: ["Bakery"], usedIn: ["Café combo"], tradable: true, tip: "One of the highest-value goods. Cake + Coffee combo at Café gives a 25% bonus." },
  { id: "coffee", icon: "☕", name: "Coffee", category: "Finished", tier: 3, basePrice: 55.0, producedBy: ["Café"], usedIn: ["Café combo"], tradable: true, tip: "Café must have Cake on the menu to unlock the combo bonus." },
  { id: "vegetable", icon: "🥦", name: "Vegetables", category: "Raw", tier: 1, basePrice: 8.0, producedBy: ["Farm"], usedIn: ["Soup","Restaurant"], tradable: true, tip: "Vegetables sell at low margin individually but are great as restaurant inputs." },
];

const WIKI_BUILDINGS = [
  { id: "farm", icon: "🌾", name: "Farm", cost: 500, difficulty: "Beginner", region: "Inland Estate", produces: ["Grain","Sugar","Vegetables"], mechanic: "Crop selection + seasonal bonuses", upgrade: "Unlock rarer crops and larger batches", desc: "The backbone of your food empire. Start here." },
  { id: "barn", icon: "🐄", name: "Barn", cost: 800, difficulty: "Beginner", region: "Inland Estate", produces: ["Dairy Milk","Steak","Eggs"], mechanic: "Animal care meter + feed cost", upgrade: "Lower feed costs and improve milk quality", desc: "Raises livestock for dairy and meat production." },
  { id: "mill", icon: "⚙️", name: "Mill", cost: 1200, difficulty: "Easy", region: "Inland Estate", produces: ["Flour"], mechanic: "Input grain → output flour with efficiency meter", upgrade: "Higher conversion ratio and faster processing", desc: "Converts raw grain into valuable flour." },
  { id: "kitchen", icon: "🍳", name: "Kitchen", cost: 1500, difficulty: "Easy", region: "New Harbor", produces: ["Dough","Butter","Cheese","Soup","Pie"], mechanic: "Recipe selector + batch size control", upgrade: "Unlock advanced recipes and larger batches", desc: "Prepares intermediate cooking ingredients." },
  { id: "bakery", icon: "🍞", name: "Bakery", cost: 2000, difficulty: "Medium", region: "New Harbor", produces: ["Bread","Cake","Cookies"], mechanic: "Ingredient checklist + quality bonus preview", upgrade: "Higher cake quality and executive bonuses", desc: "Bakes high-value finished goods for sale." },
  { id: "market_stall", icon: "🏪", name: "Market Stall", cost: 1000, difficulty: "Beginner", region: "New Harbor", produces: ["Sales Income"], mechanic: "Simple pricing with stable customer flow", upgrade: "Attract more customers per cycle", desc: "Low-risk selling for basic goods. Great for beginners." },
  { id: "cafe", icon: "☕", name: "Café", cost: 3000, difficulty: "Medium", region: "Sandy Coast", produces: ["Coffee","Cake service"], mechanic: "Menu combo builder + customer happiness meter", upgrade: "Unlock premium drinks and event bonuses", desc: "Cozy coffee shop with powerful combo mechanics." },
  { id: "food_truck", icon: "🚚", name: "Food Truck", cost: 2500, difficulty: "Medium", region: "Sandy Coast", produces: ["Sales Income"], mechanic: "Location selector + event bonus + fuel cost", upgrade: "Access exclusive event locations", desc: "Mobile high-risk high-reward sales vehicle." },
  { id: "restaurant", icon: "🍽️", name: "Restaurant", cost: 5000, difficulty: "Advanced", region: "Old Town Center", produces: ["Premium Sales Income"], mechanic: "Brand rating + menu quality + executive bonus", upgrade: "Raise brand rating and unlock VIP menus", desc: "Premium dining with the highest profit ceiling." },
  { id: "trading_hub", icon: "⚓", name: "Trading Hub", cost: 4000, difficulty: "Advanced", region: "Mountain Route", produces: ["Trade Income"], mechanic: "Auto-trading toggle + spread settings + risk display", upgrade: "Reduce fees and increase bot trade speed", desc: "Automate your market trading strategies." },
  { id: "warehouse", icon: "📦", name: "Warehouse", cost: 1500, difficulty: "Beginner", region: "Any", produces: ["+200 storage"], mechanic: "Capacity bar + upgrade tiers", upgrade: "Expand storage capacity significantly", desc: "Increases total inventory storage." },
  { id: "shop", icon: "🛒", name: "Shop", cost: 2000, difficulty: "Easy", region: "Old Town Center", produces: ["Retail Income"], mechanic: "Product shelf + demand trend + premium pricing", upgrade: "Attract higher-tier customers", desc: "Premium retail storefront for finished goods." },
];

const GUIDES = {
  "Getting Started": [
    { q: "How do I build my first farm?", a: "Go to the Map page, tap an available (blue dashed) plot, then tap 'Open Build Menu'. Choose Farm from the Production category. Costs 💰500." },
    { q: "How does production work?", a: "Select a building on the map, then tap 'Start Production' or the building's action button. Each cycle takes time and optionally consumes input resources." },
    { q: "How do I collect my goods?", a: "Go to the Collection Center (🧺) from the Town menu. Tap 'Claim All' or claim individual items. A green dot on a building means it's ready." },
    { q: "How do I sell my goods?", a: "Use the Market page to place sell orders. Or build a Market Stall, Café, or Restaurant for automatic local selling." },
  ],
  "Market Guide": [
    { q: "What are Buy and Sell Orders?", a: "A buy order is an offer to purchase at a set price. A sell order is an offer to sell at a set price. Orders fill when someone matches your price." },
    { q: "What does the K-line chart show?", a: "Candlestick (K-line) charts show price movement. Green candles = price went up. Red candles = price went down. The body shows open/close, wicks show high/low." },
    { q: "What is the Order Book?", a: "The order book lists all pending buy (green, left) and sell (red, right) orders. It shows market depth — how many units are available at each price." },
    { q: "What does price change % mean?", a: "The % shows how much the price moved from the previous period's close. +2.3% means it's 2.3% higher than the period before." },
    { q: "What is market fee?", a: "A small 0.5%–2% fee taken when your order fills. The Trading Hub research can reduce this fee." },
  ],
  "Production Guide": [
    { q: "What are input and output resources?", a: "Some buildings need input resources to produce outputs. The Mill needs Grain (input) to produce Flour (output). Check your warehouse has enough inputs." },
    { q: "What is quality?", a: "Quality % affects sell price. Premium quality (95%+) sells for more. Improved by building level, executive bonuses, and research." },
    { q: "What is a production job?", a: "A production run creates a 'job'. Completed jobs wait in the Collection Center until you claim them." },
  ],
  "Finance Guide": [
    { q: "What is the difference between Cash and Assets?", a: "Cash is spendable now. Assets include buildings, inventory, equipment — valuable but not immediately liquid." },
    { q: "How do I increase profit?", a: "Reduce costs by upgrading buildings and hiring better executives. Increase revenue by producing higher-tier goods and using the Market effectively." },
    { q: "What is the ledger?", a: "The Finance page ledger shows all your income and expense transactions, sorted by time. Green = income, Red = expense." },
  ],
};

const CATEGORIES = [
  { id: "start", label: "Getting Started", emoji: "🚀", desc: "New here? Start with this." },
  { id: "resources", label: "Resources", emoji: "🌾", desc: "Everything about goods & ingredients." },
  { id: "buildings", label: "Buildings", emoji: "🏗️", desc: "All buildings and their mechanics." },
  { id: "guides", label: "Guides", emoji: "📖", desc: "Market, production & finance guides." },
];

const RECENTLY_VIEWED = ["Farm", "Flour", "Market Guide", "Café"];

function Breadcrumb({ crumbs, onNavigate }) {
  return (
    <div className="flex items-center gap-1 text-xs text-muted-foreground mb-3 flex-wrap">
      {crumbs.map((c, i) => (
        <span key={i} className="flex items-center gap-1">
          {i > 0 && <ChevronRight className="h-3 w-3" />}
          <button
            onClick={() => onNavigate(i)}
            className={cn(i === crumbs.length - 1 ? "text-foreground font-semibold" : "hover:text-foreground transition-colors")}
          >{c}</button>
        </span>
      ))}
    </div>
  );
}

function ResourceDetail({ resource, onBack }) {
  return (
    <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }}>
      <button onClick={onBack} className="flex items-center gap-1 text-xs text-muted-foreground mb-3 hover:text-foreground transition-colors">
        <ArrowLeft className="h-3.5 w-3.5" /> Back to Resources
      </button>
      <div className="bg-card rounded-2xl border border-border p-5">
        <div className="flex items-start gap-4 mb-4">
          <div className="w-16 h-16 rounded-2xl bg-accent flex items-center justify-center text-4xl flex-shrink-0">{resource.icon}</div>
          <div className="flex-1">
            <h2 className="text-xl font-bold text-foreground">{resource.name}</h2>
            <div className="flex flex-wrap gap-1.5 mt-1.5">
              <Badge className="text-xs bg-muted text-muted-foreground border-border">Tier {resource.tier}</Badge>
              <Badge className={cn("text-xs", resource.category === "Raw" ? "bg-game-green/10 text-game-green border-game-green/20" : resource.category === "Processed" ? "bg-game-blue/10 text-game-blue border-game-blue/20" : "bg-game-purple/10 text-game-purple border-game-purple/20")}>{resource.category}</Badge>
              {resource.tradable ? <Badge className="text-xs bg-game-yellow/10 text-game-yellow border-game-yellow/20">Tradable</Badge> : <Badge variant="outline" className="text-xs text-muted-foreground">Not Tradable</Badge>}
            </div>
          </div>
          <div className="text-right">
            <p className="text-xs text-muted-foreground">Base Price</p>
            <p className="text-xl font-black text-foreground">${resource.basePrice}</p>
          </div>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-4">
          <div className="bg-muted/40 rounded-xl p-3">
            <p className="text-xs font-bold text-muted-foreground mb-1.5">🏭 Produced By</p>
            <div className="flex flex-wrap gap-1">{resource.producedBy.map(b => <Badge key={b} variant="outline" className="text-xs">{b}</Badge>)}</div>
          </div>
          <div className="bg-muted/40 rounded-xl p-3">
            <p className="text-xs font-bold text-muted-foreground mb-1.5">🍳 Used In</p>
            <div className="flex flex-wrap gap-1">{resource.usedIn.map(u => <Badge key={u} variant="outline" className="text-xs">{u}</Badge>)}</div>
          </div>
        </div>
        <div className="bg-game-yellow/5 border border-game-yellow/20 rounded-xl p-3">
          <p className="text-xs font-bold text-game-yellow mb-1">💡 Tip</p>
          <p className="text-xs text-muted-foreground">{resource.tip}</p>
        </div>
        {resource.tradable && (
          <Link to="/market">
            <Button className="w-full mt-4 bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm">
              <ExternalLink className="h-3.5 w-3.5" /> View on Market
            </Button>
          </Link>
        )}
      </div>
    </motion.div>
  );
}

function BuildingDetail({ building, onBack }) {
  const difficultyColor = { Beginner: "text-game-green", Easy: "text-game-blue", Medium: "text-game-yellow", Advanced: "text-game-red" };
  return (
    <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }}>
      <button onClick={onBack} className="flex items-center gap-1 text-xs text-muted-foreground mb-3 hover:text-foreground transition-colors">
        <ArrowLeft className="h-3.5 w-3.5" /> Back to Buildings
      </button>
      <div className="bg-card rounded-2xl border border-border p-5">
        <div className="flex items-start gap-4 mb-4">
          <div className="w-16 h-16 rounded-2xl bg-accent flex items-center justify-center text-4xl flex-shrink-0">{building.icon}</div>
          <div className="flex-1">
            <h2 className="text-xl font-bold text-foreground">{building.name}</h2>
            <p className="text-sm text-muted-foreground mt-0.5">{building.desc}</p>
            <div className="flex flex-wrap gap-1.5 mt-1.5">
              <Badge className="text-xs bg-muted text-muted-foreground border-border">💰 ${building.cost}</Badge>
              <Badge className={cn("text-xs border-transparent bg-transparent", difficultyColor[building.difficulty])}>{building.difficulty}</Badge>
              <Badge variant="outline" className="text-xs">📍 {building.region}</Badge>
            </div>
          </div>
        </div>
        <div className="space-y-3 mb-4">
          <div className="bg-muted/40 rounded-xl p-3">
            <p className="text-xs font-bold text-muted-foreground mb-1.5">📦 Produces</p>
            <div className="flex flex-wrap gap-1">{building.produces.map(p => <Badge key={p} variant="outline" className="text-xs">{p}</Badge>)}</div>
          </div>
          <div className="bg-muted/40 rounded-xl p-3">
            <p className="text-xs font-bold text-muted-foreground mb-1">⚙️ Unique Mechanic</p>
            <p className="text-xs text-foreground">{building.mechanic}</p>
          </div>
          <div className="bg-muted/40 rounded-xl p-3">
            <p className="text-xs font-bold text-muted-foreground mb-1">⬆️ Upgrade Purpose</p>
            <p className="text-xs text-foreground">{building.upgrade}</p>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2">
          <Link to="/build">
            <Button variant="outline" className="w-full gap-1.5 text-xs" size="sm">🔨 Build Page</Button>
          </Link>
          <Link to="/">
            <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-1.5 text-xs" size="sm">🗺️ View on Map</Button>
          </Link>
        </div>
      </div>
    </motion.div>
  );
}

export default function WikiPage() {
  const [view, setView] = useState("home"); // home | category | resource | building | guide
  const [category, setCategory] = useState(null);
  const [detail, setDetail] = useState(null);
  const [search, setSearch] = useState("");
  const [selectedGuide, setSelectedGuide] = useState("Getting Started");

  const searchResults = useMemo(() => {
    if (!search.trim()) return { resources: [], buildings: [] };
    const q = search.toLowerCase();
    return {
      resources: WIKI_RESOURCES.filter(r => r.name.toLowerCase().includes(q) || r.category.toLowerCase().includes(q)),
      buildings: WIKI_BUILDINGS.filter(b => b.name.toLowerCase().includes(q) || b.desc.toLowerCase().includes(q)),
    };
  }, [search]);

  const hasSearchResults = search.trim() && (searchResults.resources.length > 0 || searchResults.buildings.length > 0);

  const getBreadcrumbs = () => {
    if (view === "home") return ["Wiki"];
    if (view === "category" && category === "resources") return ["Wiki", "Resources"];
    if (view === "category" && category === "buildings") return ["Wiki", "Buildings"];
    if (view === "category" && category === "guides") return ["Wiki", "Guides"];
    if (view === "category" && category === "start") return ["Wiki", "Getting Started"];
    if (view === "resource") return ["Wiki", "Resources", detail?.name];
    if (view === "building") return ["Wiki", "Buildings", detail?.name];
    return ["Wiki"];
  };

  const handleBreadcrumb = (idx) => {
    if (idx === 0) { setView("home"); setDetail(null); }
    else if (idx === 1) { setView("category"); setDetail(null); }
  };

  return (
    <div className="p-4 max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-2 mb-4">
        <BookOpen className="h-5 w-5 text-primary" />
        <h1 className="text-xl font-bold text-foreground">Wiki</h1>
      </div>

      {/* Search */}
      <div className="relative mb-4">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input value={search} onChange={e => setSearch(e.target.value)} placeholder="Search resources, buildings, guides..." className="pl-9 h-10" />
      </div>

      {/* Breadcrumb */}
      {view !== "home" && <Breadcrumb crumbs={getBreadcrumbs()} onNavigate={handleBreadcrumb} />}

      <AnimatePresence mode="wait">
        {/* Search Results */}
        {hasSearchResults ? (
          <motion.div key="search" initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
            <p className="text-xs text-muted-foreground mb-3">Search results for "<span className="font-semibold text-foreground">{search}</span>"</p>
            {searchResults.resources.length > 0 && (
              <>
                <p className="text-xs font-bold text-muted-foreground mb-2">Resources</p>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-4">
                  {searchResults.resources.map(r => (
                    <button key={r.id} onClick={() => { setDetail(r); setView("resource"); setSearch(""); }} className="flex items-center gap-3 p-3 bg-card border border-border rounded-xl hover:shadow-sm transition-all text-left">
                      <span className="text-2xl">{r.icon}</span>
                      <div><p className="font-semibold text-sm">{r.name}</p><p className="text-xs text-muted-foreground">${r.basePrice}</p></div>
                      <ChevronRight className="h-4 w-4 text-muted-foreground ml-auto" />
                    </button>
                  ))}
                </div>
              </>
            )}
            {searchResults.buildings.length > 0 && (
              <>
                <p className="text-xs font-bold text-muted-foreground mb-2">Buildings</p>
                <div className="space-y-2">
                  {searchResults.buildings.map(b => (
                    <button key={b.id} onClick={() => { setDetail(b); setView("building"); setSearch(""); }} className="w-full flex items-center gap-3 p-3 bg-card border border-border rounded-xl hover:shadow-sm transition-all text-left">
                      <span className="text-2xl">{b.icon}</span>
                      <div className="flex-1"><p className="font-semibold text-sm">{b.name}</p><p className="text-xs text-muted-foreground">{b.desc}</p></div>
                      <ChevronRight className="h-4 w-4 text-muted-foreground" />
                    </button>
                  ))}
                </div>
              </>
            )}
          </motion.div>
        ) : view === "home" ? (
          <motion.div key="home" initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
            {/* Recently Viewed */}
            <div className="mb-4">
              <p className="text-xs font-bold text-muted-foreground mb-2">🕐 Recently Viewed</p>
              <div className="flex flex-wrap gap-1.5">
                {RECENTLY_VIEWED.map(item => (
                  <Badge key={item} variant="outline" className="text-xs cursor-pointer hover:bg-muted">{item}</Badge>
                ))}
              </div>
            </div>
            {/* Quick Start */}
            <Card className="p-4 border-game-yellow/20 bg-game-yellow/5 mb-4 cursor-pointer hover:shadow-sm transition-all" onClick={() => { setCategory("start"); setView("category"); }}>
              <div className="flex items-center gap-3">
                <span className="text-3xl">🚀</span>
                <div className="flex-1">
                  <p className="font-bold text-foreground">New Player Guide</p>
                  <p className="text-xs text-muted-foreground">Build your first Farm → Collect → Sell</p>
                </div>
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              </div>
            </Card>
            {/* Categories */}
            <div className="grid grid-cols-2 gap-3">
              {CATEGORIES.map((cat, i) => (
                <motion.div key={cat.id} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.07 }}>
                  <Card
                    className="p-4 border-border bg-card cursor-pointer hover:shadow-md transition-all"
                    onClick={() => { setCategory(cat.id); setView("category"); }}
                  >
                    <div className="text-2xl mb-2">{cat.emoji}</div>
                    <p className="font-bold text-sm text-foreground">{cat.label}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">{cat.desc}</p>
                    <ChevronRight className="h-3.5 w-3.5 text-muted-foreground mt-2" />
                  </Card>
                </motion.div>
              ))}
            </div>
          </motion.div>

        ) : view === "category" && category === "resources" ? (
          <motion.div key="resources" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }}>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {WIKI_RESOURCES.map((r, i) => (
                <motion.button key={r.id} initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.04 }}
                  onClick={() => { setDetail(r); setView("resource"); }}
                  className="flex items-center gap-3 p-3 bg-card border border-border rounded-xl hover:shadow-sm transition-all text-left"
                >
                  <span className="text-2xl">{r.icon}</span>
                  <div className="flex-1 min-w-0">
                    <p className="font-semibold text-sm text-foreground">{r.name}</p>
                    <p className="text-xs text-muted-foreground">{r.category} · ${r.basePrice}</p>
                  </div>
                  <ChevronRight className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                </motion.button>
              ))}
            </div>
          </motion.div>

        ) : view === "category" && category === "buildings" ? (
          <motion.div key="buildings" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }}>
            <div className="space-y-2">
              {WIKI_BUILDINGS.map((b, i) => (
                <motion.button key={b.id} initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.04 }}
                  onClick={() => { setDetail(b); setView("building"); }}
                  className="w-full flex items-center gap-3 p-3 bg-card border border-border rounded-xl hover:shadow-sm transition-all text-left"
                >
                  <div className="w-10 h-10 rounded-xl bg-accent flex items-center justify-center text-xl flex-shrink-0">{b.icon}</div>
                  <div className="flex-1 min-w-0">
                    <p className="font-semibold text-sm text-foreground">{b.name}</p>
                    <p className="text-xs text-muted-foreground truncate">{b.desc}</p>
                  </div>
                  <div className="text-right flex-shrink-0">
                    <p className="text-xs font-bold text-foreground">${b.cost}</p>
                    <p className="text-[10px] text-muted-foreground">{b.difficulty}</p>
                  </div>
                  <ChevronRight className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                </motion.button>
              ))}
            </div>
          </motion.div>

        ) : view === "category" && (category === "guides" || category === "start") ? (
          <motion.div key="guides" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }}>
            <div className="flex gap-3 flex-col sm:flex-row">
              <div className="sm:w-40 flex-shrink-0 space-y-1">
                {Object.keys(GUIDES).map(key => (
                  <button key={key} onClick={() => setSelectedGuide(key)}
                    className={cn("w-full text-left text-xs px-3 py-2 rounded-lg transition-all", selectedGuide === key ? "bg-accent font-semibold" : "text-muted-foreground hover:bg-muted")}
                  >{key}</button>
                ))}
              </div>
              <div className="flex-1 space-y-2">
                {GUIDES[selectedGuide]?.map((item, i) => (
                  <motion.div key={i} initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.05 }}>
                    <Card className="p-4 border-border bg-card">
                      <p className="font-bold text-sm text-foreground mb-1.5 flex items-center gap-1.5">
                        <BookOpen className="h-3.5 w-3.5 text-primary flex-shrink-0" /> {item.q}
                      </p>
                      <p className="text-xs text-muted-foreground leading-relaxed">{item.a}</p>
                    </Card>
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.div>

        ) : view === "resource" && detail ? (
          <ResourceDetail resource={detail} onBack={() => { setView("category"); setCategory("resources"); }} />
        ) : view === "building" && detail ? (
          <BuildingDetail building={detail} onBack={() => { setView("category"); setCategory("buildings"); }} />
        ) : null}
      </AnimatePresence>
    </div>
  );
}