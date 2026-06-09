import { useState } from "react";
import { ORDERS, RESOURCES } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { X, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

export default function OrdersPage() {
  const [orders, setOrders] = useState(ORDERS);

  // Group by resource
  const grouped = orders.reduce((acc, order) => {
    if (!acc[order.resource]) acc[order.resource] = [];
    acc[order.resource].push(order);
    return acc;
  }, {});

  const removeOrder = (id) => setOrders((prev) => prev.filter((o) => o.id !== id));
  const clearAll = () => setOrders([]);

  return (
    <div className="p-4 max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-bold text-foreground">📋 My Orders</h1>
        {orders.length > 0 && (
          <Button variant="outline" size="sm" className="text-game-red border-game-red/30 hover:bg-game-red/10 text-xs gap-1" onClick={clearAll}>
            <Trash2 className="h-3 w-3" /> Cancel All
          </Button>
        )}
      </div>

      {orders.length === 0 ? (
        <Card className="p-12 bg-card border-border text-center">
          <span className="text-4xl block mb-3">📭</span>
          <p className="text-sm font-medium text-muted-foreground">No active orders</p>
          <p className="text-xs text-muted-foreground mt-1">Place orders in the Market</p>
        </Card>
      ) : (
        <div className="space-y-4">
          {Object.entries(grouped).map(([resource, resourceOrders]) => {
            const r = RESOURCES[resource];
            return (
              <motion.div
                key={resource}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
              >
                <Card className="overflow-hidden bg-card border-border">
                  {/* Resource Header */}
                  <div className="flex items-center gap-2 px-4 py-3 bg-muted/40 border-b border-border/50">
                    <span className="text-xl">{r.icon}</span>
                    <h3 className="font-bold text-sm text-foreground">{r.name}</h3>
                    <Badge variant="outline" className="ml-auto text-[10px]">
                      {resourceOrders.length} order{resourceOrders.length > 1 ? "s" : ""}
                    </Badge>
                  </div>

                  {/* Orders */}
                  <div className="divide-y divide-border/50">
                    {resourceOrders.map((order) => (
                      <div key={order.id} className="flex items-center gap-3 px-4 py-3">
                        <Badge className={cn(
                          "text-[10px] px-2",
                          order.type === "buy"
                            ? "bg-game-green/10 text-game-green border-game-green/20"
                            : "bg-game-red/10 text-game-red border-game-red/20"
                        )}>
                          {order.type.toUpperCase()}
                        </Badge>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 text-xs">
                            <span className="text-muted-foreground">Remaining:</span>
                            <span className="font-bold text-foreground">{order.remaining}/{order.quantity}</span>
                          </div>
                          <div className="flex items-center gap-2 text-xs mt-0.5">
                            <span className="text-muted-foreground">Price:</span>
                            <span className="font-bold text-foreground">${order.price.toFixed(2)}</span>
                            <span className="text-muted-foreground">• {order.count} fills</span>
                          </div>
                          {/* Progress bar */}
                          <div className="w-full h-1.5 bg-muted rounded-full mt-1.5 overflow-hidden">
                            <div
                              className={cn("h-full rounded-full transition-all", order.type === "buy" ? "bg-game-green" : "bg-game-red")}
                              style={{ width: `${((order.quantity - order.remaining) / order.quantity) * 100}%` }}
                            />
                          </div>
                        </div>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7 text-muted-foreground hover:text-game-red hover:bg-game-red/10"
                          onClick={() => removeOrder(order.id)}
                        >
                          <X className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </Card>
              </motion.div>
            );
          })}
        </div>
      )}
    </div>
  );
}