import { useEffect, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { BUILDINGS } from "@/lib/gameData";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { X, ArrowUp, Move, Trash2, Play } from "lucide-react";
import { cn } from "@/lib/utils";
import BuildingDetailPanel from "./BuildingDetailPanel";

export default function MobileBottomSheet({ plot, region, onClose }) {
  const building = plot?.building ? BUILDINGS.find(b => b.id === plot.building) : null;

  return (
    <AnimatePresence>
      {plot && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="lg:hidden fixed inset-0 bg-black/20 z-40"
            onClick={onClose}
          />
          {/* Sheet */}
          <motion.div
            initial={{ y: "100%" }}
            animate={{ y: 0 }}
            exit={{ y: "100%" }}
            transition={{ type: "spring", damping: 28, stiffness: 260 }}
            className="lg:hidden fixed bottom-12 left-0 right-0 z-50 bg-card rounded-t-3xl shadow-2xl border-t border-border max-h-[75vh] overflow-y-auto"
          >
            {/* Handle */}
            <div className="flex justify-center pt-3 pb-1">
              <div className="w-10 h-1 bg-border rounded-full" />
            </div>

            <div className="px-4 pb-6 pt-1">
              {plot.state === "available" ? (
                <div className="text-center py-2">
                  <span className="text-4xl block mb-2">🏗️</span>
                  <h3 className="font-bold text-lg text-foreground mb-1">Build Here</h3>
                  <p className="text-xs text-muted-foreground mb-3">This plot is available for construction.</p>
                  {region && (
                    <>
                      <p className="text-xs font-semibold text-muted-foreground mb-2">Best for {region.name}:</p>
                      <div className="flex flex-wrap gap-1.5 justify-center mb-3">
                        {region.suggested.map(s => <Badge key={s} variant="outline" className="text-xs">{s}</Badge>)}
                      </div>
                      <Badge className="bg-game-blue/10 text-game-blue border-game-blue/20 text-xs mb-4">{region.bonus}</Badge>
                    </>
                  )}
                  <Button className="w-full bg-game-blue hover:bg-game-blue/90 text-white rounded-full h-12 font-bold" onClick={onClose}>
                    🔨 Open Build Menu
                  </Button>
                </div>
              ) : plot.state === "locked" ? (
                <div className="text-center py-2">
                  <span className="text-4xl block mb-2">🔒</span>
                  <h3 className="font-bold text-foreground">Locked Plot</h3>
                  <p className="text-xs text-muted-foreground mt-1">Unlock more plots by leveling up.</p>
                  <Button variant="outline" className="mt-4 w-full rounded-full" onClick={onClose}>Close</Button>
                </div>
              ) : (
                <BuildingDetailPanel plot={plot} onClose={onClose} />
              )}
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}