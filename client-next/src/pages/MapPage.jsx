import { useState } from "react";
import { MAP_PLOTS, BUILDINGS, RESOURCES } from "@/lib/gameData";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Lock, ChevronLeft, ChevronRight, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";
import BuildingDetailPanel from "@/components/game/BuildingDetailPanel";
import MobileBottomSheet from "@/components/game/MobileBottomSheet";

const REGIONS = [
  {
    id: "harbor", name: "New Harbor", icon: "⚓", emoji: "🌊",
    bg: "from-blue-100/60 to-cyan-50/40",
    deco: ["🌊","🚢","🌊","🐟","🌊","⛵","🌊"],
    decoBottom: ["🌊","🐙","🌊","🦀","🌊","🐠","🌊"],
    slots: "4/12", locked: false,
    desc: "Coastal harbor town — great for trading & early production",
    bonus: "+5% Trading Speed",
    suggested: ["Farm","Market Stall","Trading Hub","Warehouse"],
    color: "text-blue-600",
  },
  {
    id: "inland", name: "Inland Estate", icon: "🌾", emoji: "🌿",
    bg: "from-green-100/60 to-lime-50/40",
    deco: ["🌿","🐄","🌾","🌻","🌾","🐑","🌿"],
    decoBottom: ["🌾","🐇","🌿","🦋","🌾","🐝","🌿"],
    slots: "2/10", locked: false,
    desc: "Rich farmland — best for crops, barns, and mills",
    bonus: "+8% Crop Output",
    suggested: ["Farm","Barn","Mill","Warehouse"],
    color: "text-green-700",
  },
  {
    id: "coast", name: "Sandy Coast", icon: "🏖️", emoji: "☀️",
    bg: "from-yellow-100/60 to-orange-50/40",
    deco: ["🏖️","🌴","☀️","🦀","🌊","🍹","🏄"],
    decoBottom: ["🐚","🦞","🌊","🏄","🐬","⛱️","🌊"],
    slots: "0/8", locked: true, unlockReq: "Level 20",
    desc: "Warm beach market — tourism, cafes, food trucks",
    bonus: "+12% Food Sales",
    suggested: ["Café","Food Truck","Market Stall","Shop"],
    color: "text-orange-600",
  },
  {
    id: "mountain", name: "Mountain Route", icon: "⛰️", emoji: "🏔️",
    bg: "from-slate-100/60 to-stone-50/40",
    deco: ["⛰️","🦅","🌲","🐻","🌲","❄️","⛰️"],
    decoBottom: ["🌲","🦌","⛰️","🌿","🏔️","🦌","🌲"],
    slots: "0/8", locked: true, unlockReq: "Level 30",
    desc: "Mountain trade road — premium shops & high-risk markets",
    bonus: "+15% Trade Margin",
    suggested: ["Trading Hub","Restaurant","Shop","Warehouse"],
    color: "text-slate-600",
  },
  {
    id: "oldtown", name: "Old Town Center", icon: "🏙️", emoji: "🏛️",
    bg: "from-purple-100/40 to-rose-50/30",
    deco: ["🏛️","🎭","🏙️","🎪","🏗️","🎨","🏛️"],
    decoBottom: ["🎭","🎪","🏙️","🎨","🛍️","🏛️","🎭"],
    slots: "0/10", locked: true, unlockReq: "Level 40",
    desc: "Dense city streets — restaurants, finance, social hubs",
    bonus: "+20% Brand Income",
    suggested: ["Restaurant","Shop","Café","Finance"],
    color: "text-purple-600",
  },
];

const BUILDING_VISUALS = {
  farm:         { bg: "bg-green-100", border: "border-green-300", statusColor: "bg-game-green" },
  barn:         { bg: "bg-amber-100", border: "border-amber-300", statusColor: "bg-amber-400" },
  mill:         { bg: "bg-stone-100", border: "border-stone-300", statusColor: "bg-stone-400" },
  kitchen:      { bg: "bg-orange-100", border: "border-orange-300", statusColor: "bg-orange-400" },
  bakery:       { bg: "bg-pink-100", border: "border-pink-300", statusColor: "bg-pink-400" },
  market_stall: { bg: "bg-yellow-100", border: "border-yellow-300", statusColor: "bg-yellow-400" },
  cafe:         { bg: "bg-amber-100", border: "border-amber-400", statusColor: "bg-amber-500" },
  food_truck:   { bg: "bg-teal-100", border: "border-teal-300", statusColor: "bg-teal-400" },
  restaurant:   { bg: "bg-red-100", border: "border-red-300", statusColor: "bg-red-400" },
  trading_hub:  { bg: "bg-blue-100", border: "border-blue-300", statusColor: "bg-blue-400" },
  warehouse:    { bg: "bg-gray-100", border: "border-gray-300", statusColor: "bg-gray-400" },
  shop:         { bg: "bg-violet-100", border: "border-violet-300", statusColor: "bg-violet-400" },
};

const STATUS_INDICATORS = {
  farm: "🌱 Growing",
  barn: "🐄 Producing",
  mill: "⚙️ Processing",
  kitchen: "🍳 Cooking",
  bakery: "🔥 Baking",
  market_stall: "🟢 Open",
  cafe: "☕ Serving",
  food_truck: "🚗 Driving",
  restaurant: "🍽️ Dining",
  trading_hub: "📊 Trading",
  warehouse: "📦 Storing",
  shop: "🛍️ Selling",
};

function PlotTile({ plot, isSelected, onClick }) {
  const building = plot.building ? BUILDINGS.find(b => b.id === plot.building) : null;
  const vis = building ? (BUILDING_VISUALS[building.id] || { bg: "bg-card", border: "border-border" }) : null;
  const isReady = plot.id % 3 === 0 && plot.state === "occupied";

  return (
    <motion.button
      whileHover={{ scale: 1.04, y: -2 }}
      whileTap={{ scale: 0.96 }}
      onClick={() => onClick(plot)}
      className={cn(
        "relative aspect-square rounded-2xl border-2 transition-all flex flex-col items-center justify-center gap-0.5 p-1 shadow-sm",
        plot.state === "occupied" && vis && `${vis.bg} ${vis.border} hover:shadow-md`,
        plot.state === "available" && "bg-game-blue/10 border-game-blue/30 border-dashed hover:bg-game-blue/20 hover:border-game-blue/50",
        plot.state === "locked" && "bg-muted/50 border-border/40 cursor-not-allowed opacity-50",
        isSelected && "ring-2 ring-game-blue ring-offset-2 ring-offset-background shadow-lg"
      )}
    >
      {plot.state === "occupied" && building && (
        <>
          {isReady && (
            <span className="absolute -top-1 -right-1 w-4 h-4 bg-game-green rounded-full border-2 border-white animate-pulse flex items-center justify-center text-[8px]">!</span>
          )}
          <span className="text-2xl sm:text-3xl drop-shadow-sm">{building.icon}</span>
          <span className="text-[9px] sm:text-[10px] font-bold text-foreground truncate w-full text-center leading-tight">{building.name}</span>
          <Badge className="text-[7px] px-1 py-0 h-3.5 bg-white/80 text-foreground border-border/50 shadow-sm">
            Lv.{plot.level}
          </Badge>
        </>
      )}
      {plot.state === "available" && (
        <div className="flex flex-col items-center">
          <span className="text-2xl text-game-blue/50">＋</span>
          <span className="text-[9px] text-game-blue/70 font-semibold">Build</span>
        </div>
      )}
      {plot.state === "locked" && (
        <Lock className="h-5 w-5 text-muted-foreground/40" />
      )}
    </motion.button>
  );
}

export default function MapPage() {
  const [selectedPlot, setSelectedPlot] = useState(null);
  const [regionIdx, setRegionIdx] = useState(0);
  const region = REGIONS[regionIdx];

  const handlePlotClick = (plot) => {
    setSelectedPlot(plot.id === selectedPlot?.id ? null : plot);
  };

  return (
    <div className="relative flex flex-col lg:flex-row h-[calc(100vh-3.5rem-2.5rem)]">
      {/* Map Area */}
      <div className={cn("flex-1 overflow-auto bg-gradient-to-br transition-all duration-700", region.bg)}>
        <div className="p-3 pb-0">
          {/* Region Selector */}
          <div className="overflow-x-auto pb-2 mb-2">
            <div className="flex gap-2 min-w-max">
              {REGIONS.map((r, i) => (
                <button
                  key={r.id}
                  onClick={() => !r.locked && setRegionIdx(i)}
                  disabled={r.locked}
                  className={cn(
                    "flex items-center gap-2 px-3 py-2 rounded-xl border text-xs font-medium transition-all flex-shrink-0",
                    r.locked && "opacity-40 cursor-not-allowed",
                    regionIdx === i && !r.locked
                      ? "bg-white border-primary shadow-md text-foreground"
                      : "bg-white/60 border-white/80 text-muted-foreground hover:bg-white/80"
                  )}
                >
                  <span className="text-base">{r.icon}</span>
                  <div className="text-left hidden sm:block">
                    <p className="font-bold text-[11px]">{r.name}</p>
                    <p className="text-[9px] text-muted-foreground">{r.locked ? `🔒 ${r.unlockReq}` : r.slots + " buildings"}</p>
                  </div>
                  <span className="sm:hidden font-semibold">{r.name.split(" ")[0]}</span>
                  {r.locked && <Lock className="h-3 w-3 flex-shrink-0" />}
                </button>
              ))}
            </div>
          </div>

          {/* Region Info */}
          <AnimatePresence mode="wait">
            <motion.div
              key={region.id}
              initial={{ opacity: 0, y: -4 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              className="flex items-center gap-3 mb-3 bg-white/60 backdrop-blur-sm rounded-xl px-3 py-2 border border-white/80"
            >
              <span className="text-2xl">{region.icon}</span>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-bold text-sm text-foreground">{region.name}</span>
                  <Badge className={cn("text-[10px] bg-white/80 border-border/40", region.color)}>{region.bonus}</Badge>
                  <Badge variant="outline" className="text-[10px]">{region.slots} slots</Badge>
                </div>
                <p className="text-xs text-muted-foreground truncate">{region.desc}</p>
              </div>
            </motion.div>
          </AnimatePresence>
        </div>

        {/* Map Content */}
        <AnimatePresence mode="wait">
          <motion.div
            key={region.id}
            initial={{ opacity: 0, scale: 0.97 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.3 }}
            className="px-3"
          >
            <div className="text-center text-xl mb-2 opacity-50 tracking-widest">
              {region.deco.join("")}
            </div>
            <div className="max-w-sm mx-auto">
              <div className="grid grid-cols-4 gap-2 sm:gap-3">
                {MAP_PLOTS.map((plot) => (
                  <PlotTile
                    key={plot.id}
                    plot={plot}
                    isSelected={selectedPlot?.id === plot.id}
                    onClick={handlePlotClick}
                  />
                ))}
              </div>
            </div>
            <div className="text-center text-xl mt-2 opacity-50 tracking-widest">
              {region.decoBottom.join("")}
            </div>

            {/* Legend */}
            <div className="flex flex-wrap gap-2 mt-3 justify-center text-[10px] text-muted-foreground pb-4">
              <span className="flex items-center gap-1 bg-white/60 px-2 py-1 rounded-full"><span className="w-2.5 h-2.5 rounded bg-green-100 border border-green-300 inline-block" /> Occupied</span>
              <span className="flex items-center gap-1 bg-white/60 px-2 py-1 rounded-full"><span className="w-2.5 h-2.5 rounded bg-game-blue/10 border border-dashed border-game-blue/30 inline-block" /> Available</span>
              <span className="flex items-center gap-1 bg-white/60 px-2 py-1 rounded-full"><span className="w-2.5 h-2.5 rounded bg-muted/50 border border-border/40 inline-block" /> Locked</span>
              <span className="flex items-center gap-1 bg-white/60 px-2 py-1 rounded-full"><span className="w-2.5 h-2.5 rounded-full bg-game-green inline-block" /> Ready</span>
            </div>
          </motion.div>
        </AnimatePresence>
      </div>

      {/* Desktop Detail Panel */}
      <div className="hidden lg:block w-72 border-l border-border bg-card/60 backdrop-blur-sm overflow-y-auto">
        <AnimatePresence mode="wait">
          {selectedPlot && selectedPlot.state === "occupied" ? (
            <div className="p-4">
              <BuildingDetailPanel
                key={selectedPlot.id}
                plot={selectedPlot}
                onClose={() => setSelectedPlot(null)}
              />
            </div>
          ) : selectedPlot && selectedPlot.state === "available" ? (
            <motion.div key="available" initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="p-4">
              <div className="bg-game-blue/5 border border-game-blue/20 rounded-2xl p-4 text-center">
                <span className="text-3xl block mb-2">🏗️</span>
                <h3 className="font-bold text-foreground mb-1">Empty Plot</h3>
                <p className="text-xs text-muted-foreground mb-3">Build a structure here to start producing goods.</p>
                <p className="text-xs font-semibold text-muted-foreground mb-2">Suggested for {region.name}:</p>
                <div className="flex flex-wrap gap-1 justify-center mb-3">
                  {region.suggested.map(s => <Badge key={s} variant="outline" className="text-[10px]">{s}</Badge>)}
                </div>
                <Badge className="bg-game-blue/10 text-game-blue border-game-blue/20 text-xs">{region.bonus}</Badge>
              </div>
            </motion.div>
          ) : (
            <motion.div key="empty" initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="h-full flex flex-col items-center justify-center text-center text-muted-foreground py-12 p-4">
              <span className="text-4xl mb-3">{region.emoji}</span>
              <p className="text-sm font-medium">{region.name}</p>
              <p className="text-xs mt-1 text-muted-foreground">{region.desc}</p>
              <Badge className={cn("mt-3 text-xs", region.color)}>{region.bonus}</Badge>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Mobile Bottom Sheet */}
      <MobileBottomSheet
        plot={selectedPlot}
        region={region}
        onClose={() => setSelectedPlot(null)}
      />
    </div>
  );
}