import { useState } from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { CheckCircle2, Clock, AlertTriangle, Zap, Package } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";
import { RESOURCES } from "@/lib/gameData";

const PRODUCTION_JOBS = [
  { id: 1, buildingIcon: "🌾", buildingName: "Farm (Plot #1)", resource: "wheat", qty: 120, quality: 92, completedAt: "2m ago", status: "ready" },
  { id: 2, buildingIcon: "🍞", buildingName: "Bakery (Plot #6)", resource: "bread", qty: 30, quality: 97, completedAt: "5m ago", status: "ready" },
  { id: 3, buildingIcon: "🐄", buildingName: "Barn (Plot #2)", resource: "milk", qty: 50, quality: 95, completedAt: "Just now", status: "ready" },
  { id: 4, buildingIcon: "⚙️", buildingName: "Mill (Plot #5)", resource: "flour", qty: 80, quality: 88, completedAt: "1m ago", status: "ready" },
  { id: 5, buildingIcon: "☕", buildingName: "Café (Plot #11)", resource: "coffee", qty: 15, quality: 93, completedAt: "10m ago", status: "ready" },
  { id: 6, buildingIcon: "🌾", buildingName: "Farm (Plot #1)", resource: "corn", qty: 200, quality: 85, remainingTime: "12m", progress: 72, status: "in_progress" },
  { id: 7, buildingIcon: "🍳", buildingName: "Kitchen (Plot #8)", resource: "pie", qty: 8, quality: 90, remainingTime: "28m", progress: 45, status: "in_progress" },
  { id: 8, buildingIcon: "🐄", buildingName: "Barn (Plot #2)", resource: "egg", qty: 60, quality: 88, remainingTime: "1h 5m", progress: 20, status: "in_progress" },
  { id: 9, buildingIcon: "🏪", buildingName: "Market Stall (Plot #10)", resource: "bread", qty: 0, quality: 0, status: "idle" },
  { id: 10, buildingIcon: "📦", buildingName: "Warehouse (Plot #9)", resource: "wheat", qty: 340, quality: 92, status: "full", warehouseFull: true },
];

const qualityColor = (q) => {
  if (q >= 95) return "text-game-purple";
  if (q >= 85) return "text-game-green";
  if (q >= 70) return "text-game-yellow";
  return "text-game-red";
};

const qualityLabel = (q) => {
  if (q >= 95) return "⭐ Premium";
  if (q >= 85) return "✅ Good";
  if (q >= 70) return "🟡 Fair";
  return "⚠️ Poor";
};

function ReadyCard({ job, onClaim }) {
  const res = RESOURCES[job.resource];
  return (
    <motion.div layout initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 0.9 }}>
      <Card className="p-4 bg-card border-game-green/20 border hover:shadow-md transition-all">
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 rounded-xl bg-game-green/10 flex items-center justify-center text-2xl flex-shrink-0">
            {job.buildingIcon}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-xs text-muted-foreground truncate">{job.buildingName}</p>
            <div className="flex items-center gap-2 mt-0.5">
              <span className="text-lg">{res?.icon}</span>
              <span className="font-bold text-sm text-foreground">{job.qty} {res?.name}</span>
            </div>
            <div className="flex items-center gap-2 mt-1">
              <span className={cn("text-[10px] font-semibold", qualityColor(job.quality))}>{qualityLabel(job.quality)}</span>
              <span className="text-[10px] text-muted-foreground">· {job.completedAt}</span>
            </div>
          </div>
          <Button
            onClick={() => onClaim(job.id)}
            size="sm"
            className="bg-game-blue hover:bg-game-blue/90 text-white rounded-full px-4 flex-shrink-0"
          >
            Claim
          </Button>
        </div>
      </Card>
    </motion.div>
  );
}

function InProgressCard({ job }) {
  const res = RESOURCES[job.resource];
  return (
    <Card className="p-4 bg-card border-border">
      <div className="flex items-center gap-3">
        <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center text-2xl flex-shrink-0">
          {job.buildingIcon}
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-xs text-muted-foreground truncate">{job.buildingName}</p>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="text-base">{res?.icon}</span>
            <span className="font-bold text-sm text-foreground">{job.qty} {res?.name}</span>
          </div>
          <div className="mt-2">
            <div className="flex justify-between text-[10px] text-muted-foreground mb-1">
              <span>⏱ {job.remainingTime} left</span>
              <span>{job.progress}%</span>
            </div>
            <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
              <div className="h-full bg-game-blue rounded-full transition-all" style={{ width: `${job.progress}%` }} />
            </div>
          </div>
        </div>
        <Button variant="outline" size="sm" className="text-[10px] h-8 px-2 flex-shrink-0 gap-1 text-game-yellow border-game-yellow/30 hover:bg-game-yellow/10">
          <Zap className="h-3 w-3" /> Speed
        </Button>
      </div>
    </Card>
  );
}

export default function CollectionPage() {
  const [jobs, setJobs] = useState(PRODUCTION_JOBS);
  const [claimedSummary, setClaimedSummary] = useState(null);
  const [tab, setTab] = useState("ready");

  const readyJobs = jobs.filter(j => j.status === "ready");
  const inProgressJobs = jobs.filter(j => j.status === "in_progress");
  const idleJobs = jobs.filter(j => j.status === "idle");
  const fullJobs = jobs.filter(j => j.status === "full");
  const warehouseNearFull = fullJobs.length > 0;

  const handleClaim = (id) => {
    const job = jobs.find(j => j.id === id);
    setJobs(prev => prev.filter(j => j.id !== id));
    setClaimedSummary([{ resource: job.resource, qty: job.qty }]);
    setTimeout(() => setClaimedSummary(null), 3000);
  };

  const handleClaimAll = () => {
    const ready = jobs.filter(j => j.status === "ready");
    const summary = ready.reduce((acc, j) => {
      const existing = acc.find(a => a.resource === j.resource);
      if (existing) existing.qty += j.qty;
      else acc.push({ resource: j.resource, qty: j.qty });
      return acc;
    }, []);
    setJobs(prev => prev.filter(j => j.status !== "ready"));
    setClaimedSummary(summary);
    setTimeout(() => setClaimedSummary(null), 4000);
  };

  return (
    <div className="p-4 max-w-2xl mx-auto pb-24">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-xl font-bold text-foreground">🧺 Collection</h1>
          <p className="text-xs text-muted-foreground">{readyJobs.length} items ready to collect</p>
        </div>
        {readyJobs.length > 0 && (
          <Button
            onClick={handleClaimAll}
            className="bg-game-blue hover:bg-game-blue/90 text-white gap-2 rounded-full px-5 font-bold shadow-md"
          >
            <CheckCircle2 className="h-4 w-4" /> Claim All ({readyJobs.length})
          </Button>
        )}
      </div>

      {/* Warehouse Warning */}
      {warehouseNearFull && (
        <div className="mb-4 flex items-start gap-2.5 bg-game-yellow/10 border border-game-yellow/30 rounded-xl p-3">
          <AlertTriangle className="h-4 w-4 text-game-yellow flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-xs font-bold text-game-yellow">Warehouse Full!</p>
            <p className="text-xs text-muted-foreground">Upgrade your warehouse or sell resources before collecting.</p>
          </div>
        </div>
      )}

      {/* Claimed Summary Toast */}
      <AnimatePresence>
        {claimedSummary && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="mb-4 bg-game-green/10 border border-game-green/30 rounded-xl p-3"
          >
            <p className="text-xs font-bold text-game-green mb-1.5">✅ Collected!</p>
            <div className="flex flex-wrap gap-1.5">
              {claimedSummary.map((item) => (
                <span key={item.resource} className="bg-game-green/10 text-game-green text-xs font-medium px-2 py-0.5 rounded-full border border-game-green/20">
                  {RESOURCES[item.resource]?.icon} +{item.qty} {RESOURCES[item.resource]?.name}
                </span>
              ))}
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Tabs */}
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="w-full bg-muted/50 mb-4">
          <TabsTrigger value="ready" className="flex-1 text-xs gap-1">
            ✅ Ready {readyJobs.length > 0 && <Badge className="bg-game-green/20 text-game-green border-none text-[10px] px-1.5 h-4 min-w-4">{readyJobs.length}</Badge>}
          </TabsTrigger>
          <TabsTrigger value="in_progress" className="flex-1 text-xs gap-1">
            ⏳ In Progress {inProgressJobs.length > 0 && <Badge className="bg-game-blue/20 text-game-blue border-none text-[10px] px-1.5 h-4 min-w-4">{inProgressJobs.length}</Badge>}
          </TabsTrigger>
          <TabsTrigger value="idle" className="flex-1 text-xs">
            💤 Idle
          </TabsTrigger>
        </TabsList>

        <TabsContent value="ready">
          <AnimatePresence mode="popLayout">
            {readyJobs.length === 0 ? (
              <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-center py-12 text-muted-foreground">
                <span className="text-4xl block mb-2">🎉</span>
                <p className="text-sm font-medium">All collected!</p>
                <p className="text-xs mt-1">Check back later for more production.</p>
              </motion.div>
            ) : (
              <div className="space-y-2">
                {readyJobs.map(job => <ReadyCard key={job.id} job={job} onClaim={handleClaim} />)}
              </div>
            )}
          </AnimatePresence>
        </TabsContent>

        <TabsContent value="in_progress">
          <div className="space-y-2">
            {inProgressJobs.length === 0 ? (
              <div className="text-center py-12 text-muted-foreground">
                <span className="text-4xl block mb-2">🏗️</span>
                <p className="text-sm font-medium">Nothing in production</p>
                <p className="text-xs mt-1">Start producing goods from your buildings.</p>
              </div>
            ) : (
              inProgressJobs.map(job => <InProgressCard key={job.id} job={job} />)
            )}
          </div>
        </TabsContent>

        <TabsContent value="idle">
          <div className="space-y-2">
            {idleJobs.length === 0 ? (
              <div className="text-center py-12 text-muted-foreground">
                <Package className="h-8 w-8 mx-auto mb-2 opacity-30" />
                <p className="text-sm font-medium">All buildings are active!</p>
              </div>
            ) : (
              idleJobs.map(job => (
                <Card key={job.id} className="p-4 border-border bg-muted/30">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center text-2xl">{job.buildingIcon}</div>
                    <div>
                      <p className="font-medium text-sm text-foreground">{job.buildingName}</p>
                      <p className="text-xs text-muted-foreground">Idle — no production started</p>
                    </div>
                    <Button size="sm" className="ml-auto bg-game-blue hover:bg-game-blue/90 text-white text-xs rounded-full">
                      Start
                    </Button>
                  </div>
                </Card>
              ))
            )}
          </div>
        </TabsContent>
      </Tabs>

      {/* Fixed bottom bar on mobile */}
      <div className="fixed bottom-16 md:bottom-10 left-0 right-0 md:hidden px-4 pb-2">
        {readyJobs.length > 0 && (
          <Button
            onClick={handleClaimAll}
            className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2 rounded-full h-12 font-bold shadow-xl text-base"
          >
            <CheckCircle2 className="h-5 w-5" /> Claim All ({readyJobs.length})
          </Button>
        )}
      </div>
    </div>
  );
}