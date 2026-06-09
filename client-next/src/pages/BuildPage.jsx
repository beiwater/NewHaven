import { useState } from "react";
import { BUILDINGS, RESOURCES } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

const CATEGORIES = [
  { id: "all", label: "All", emoji: "📋" },
  { id: "production", label: "Production", emoji: "🌾" },
  { id: "processing", label: "Processing", emoji: "⚙️" },
  { id: "commerce", label: "Commerce", emoji: "🏪" },
  { id: "storage", label: "Storage", emoji: "📦" },
];

export default function BuildPage() {
  const [category, setCategory] = useState("all");

  const filtered = category === "all"
    ? BUILDINGS
    : BUILDINGS.filter((b) => b.category === category);

  return (
    <div className="p-4 max-w-4xl mx-auto">
      <div className="flex items-center gap-2 mb-4">
        <h1 className="text-xl font-bold text-foreground">🔨 Building Shop</h1>
      </div>

      {/* Category Tabs */}
      <Tabs value={category} onValueChange={setCategory} className="mb-4">
        <TabsList className="bg-muted/50 h-auto flex-wrap">
          {CATEGORIES.map((c) => (
            <TabsTrigger key={c.id} value={c.id} className="text-xs gap-1 data-[state=active]:bg-accent">
              <span>{c.emoji}</span> {c.label}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {/* Building Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {filtered.map((building, i) => (
          <motion.div
            key={building.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.05 }}
          >
            <Card className="p-4 bg-card hover:shadow-md transition-shadow border-border">
              <div className="flex items-start gap-3">
                <div className="w-12 h-12 rounded-xl bg-accent flex items-center justify-center text-2xl flex-shrink-0">
                  {building.icon}
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-bold text-sm text-foreground">{building.name}</h3>
                  <p className="text-xs text-muted-foreground mt-0.5">{building.desc}</p>
                </div>
              </div>

              {building.produces.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-3">
                  {building.produces.map((r) => (
                    <Badge key={r} variant="outline" className="text-[10px] gap-1 px-1.5 py-0">
                      {RESOURCES[r]?.icon} {RESOURCES[r]?.name}
                    </Badge>
                  ))}
                </div>
              )}

              <div className="flex items-center justify-between mt-3 pt-3 border-t border-border/50">
                <div className="flex items-center gap-1">
                  <span className="text-sm">💰</span>
                  <span className="text-sm font-bold text-foreground">${building.price.toLocaleString()}</span>
                </div>
                <Button size="sm" className="bg-primary hover:bg-primary/90 text-primary-foreground text-xs px-4 rounded-full">
                  Buy
                </Button>
              </div>
            </Card>
          </motion.div>
        ))}
      </div>
    </div>
  );
}