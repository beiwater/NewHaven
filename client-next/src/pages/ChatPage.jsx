import { useState } from "react";
import { CHAT_MESSAGES } from "@/lib/gameData";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Send, Image, Hash, Search, ArrowLeft, Circle } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";

const CHANNELS = [
  { id: "general", label: "General", emoji: "💬" },
  { id: "sales", label: "Sales", emoji: "🏷️" },
  { id: "help", label: "Help", emoji: "❓" },
];

const DM_CONTACTS = [
  { id: 1, name: "Atlas Trading Bot", avatar: "🤖", role: "Trade Assistant", online: true, lastMsg: "Wheat is rising! Buy now before it peaks 📈", time: "2m" },
  { id: 2, name: "Nova Market Bot", avatar: "🌟", role: "Market Analyst", online: true, lastMsg: "Your sell order for Fish filled at $19.0", time: "5m" },
  { id: 3, name: "Harbor Cafe Owner", avatar: "☕", role: "Player", online: true, lastMsg: "Hey! Can you supply 20 Coffee next week?", time: "12m" },
  { id: 4, name: "Inland Farmer", avatar: "👨🌾", role: "Player", online: false, lastMsg: "I have 200 Grain for sale at $11.8 each", time: "1h" },
  { id: 5, name: "Restaurant Manager", avatar: "🍽️", role: "Player", online: false, lastMsg: "What's your Cake price? Need 30 units ASAP", time: "3h" },
];

const DM_MESSAGES = {
  1: [
    { id: 1, from: "Atlas Trading Bot", avatar: "🤖", msg: "Good morning! Wheat prices are trending upward today 📈", time: "9:00", isMe: false },
    { id: 2, from: "Atlas Trading Bot", avatar: "🤖", msg: "Current buy price: $12.8 | Sell price: $12.2 | 24h change: +2.3%", time: "9:01", isMe: false },
    { id: 3, from: "Me", avatar: "🧑🍳", msg: "Thanks! Should I buy more?", time: "9:05", isMe: true },
    { id: 4, from: "Atlas Trading Bot", avatar: "🤖", msg: "Based on recent trends, yes! The 7-day average suggests a continued rise. Consider buying before 11:00 AM.", time: "9:05", isMe: false },
  ],
  2: [
    { id: 1, from: "Nova Market Bot", avatar: "🌟", msg: "Your sell order for 30x Fish at $19.0 has been fully filled ✅", time: "10:30", isMe: false },
    { id: 2, from: "Nova Market Bot", avatar: "🌟", msg: "Total revenue: $570.0 after 0.5% market fee", time: "10:30", isMe: false },
    { id: 3, from: "Me", avatar: "🧑🍳", msg: "Great! Place another sell order for 50 fish at 18.5", time: "10:32", isMe: true },
    { id: 4, from: "Nova Market Bot", avatar: "🌟", msg: "Order placed! 50x Fish @ $18.5. Estimated fill time: ~2 hours based on current market activity.", time: "10:32", isMe: false },
  ],
  3: [
    { id: 1, from: "Harbor Cafe Owner", avatar: "☕", msg: "Hey Captain Mochi! 👋 I run the harbor café and I'm looking for a coffee supplier", time: "Yesterday", isMe: false },
    { id: 2, from: "Me", avatar: "🧑🍳", msg: "Hi! Yes I produce coffee. What quantity do you need?", time: "Yesterday", isMe: true },
    { id: 3, from: "Harbor Cafe Owner", avatar: "☕", msg: "Around 20 units weekly. Can you do $53 per unit?", time: "Yesterday", isMe: false },
    { id: 4, from: "Me", avatar: "🧑🍳", msg: "That works for me! Let's set up a weekly contract 📜", time: "Yesterday", isMe: true },
    { id: 5, from: "Harbor Cafe Owner", avatar: "☕", msg: "Perfect! Also interested in Cake if you have it 🎂", time: "12m", isMe: false },
  ],
  4: [
    { id: 1, from: "Inland Farmer", avatar: "👨🌾", msg: "Hi! I have a large Grain harvest — 200 units at $11.8 each. Interested?", time: "1h", isMe: false },
    { id: 2, from: "Me", avatar: "🧑🍳", msg: "That's a good deal! Let me check my warehouse capacity", time: "58m", isMe: true },
    { id: 3, from: "Inland Farmer", avatar: "👨🌾", msg: "No rush, offer stands for 24h 🌾", time: "57m", isMe: false },
  ],
  5: [
    { id: 1, from: "Restaurant Manager", avatar: "🍽️", msg: "Hello! I urgently need 30 Cakes. What's your price per unit?", time: "3h", isMe: false },
    { id: 2, from: "Me", avatar: "🧑🍳", msg: "Current production price is $67 each. I can do 25 units now", time: "2h", isMe: true },
    { id: 3, from: "Restaurant Manager", avatar: "🍽️", msg: "I'll take the 25. Can you also supply regularly? We host events every week 🎪", time: "2h", isMe: false },
  ],
};

function ChatBubble({ msg }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 5 }}
      animate={{ opacity: 1, y: 0 }}
      className={cn("flex gap-2 max-w-[85%]", msg.isMe ? "ml-auto flex-row-reverse" : "")}
    >
      <span className="text-xl flex-shrink-0 mt-1">{msg.avatar}</span>
      <div>
        <div className={cn("flex items-center gap-2 mb-0.5", msg.isMe ? "flex-row-reverse" : "")}>
          <span className="text-xs font-bold text-foreground">{msg.isMe ? "You" : (msg.user || msg.from)}</span>
          <span className="text-[10px] text-muted-foreground">{msg.time}</span>
        </div>
        <div className={cn(
          "rounded-2xl px-3 py-2 text-sm",
          msg.isMe ? "bg-game-blue text-white rounded-tr-sm" : "bg-card border border-border rounded-tl-sm"
        )}>
          {msg.isImage ? (
            <div className="w-40 h-28 bg-muted/50 rounded-lg flex items-center justify-center">
              <div className="text-center"><span className="text-3xl">🎂</span><p className="text-xs text-muted-foreground mt-1">cake_showcase.jpg</p></div>
            </div>
          ) : <p className="leading-relaxed">{msg.msg}</p>}
        </div>
      </div>
    </motion.div>
  );
}

export default function ChatPage() {
  const [tab, setTab] = useState("public");
  const [channel, setChannel] = useState("general");
  const [message, setMessage] = useState("");
  const [search, setSearch] = useState("");
  const [activeDM, setActiveDM] = useState(null);
  const [dmMessage, setDmMessage] = useState("");
  const [localMessages, setLocalMessages] = useState(DM_MESSAGES);

  const filteredMessages = CHAT_MESSAGES.filter(m => m.channel === channel);
  const filteredContacts = DM_CONTACTS.filter(c => c.name.toLowerCase().includes(search.toLowerCase()));
  const currentContact = DM_CONTACTS.find(c => c.id === activeDM);
  const currentDMs = activeDM ? (localMessages[activeDM] || []) : [];

  const sendDM = () => {
    if (!dmMessage.trim() || !activeDM) return;
    const newMsg = { id: Date.now(), from: "Me", avatar: "🧑🍳", msg: dmMessage, time: "now", isMe: true };
    setLocalMessages(prev => ({ ...prev, [activeDM]: [...(prev[activeDM] || []), newMsg] }));
    setDmMessage("");
  };

  return (
    <div className="flex flex-col h-[calc(100vh-3.5rem-2.5rem)]">
      {/* Tab bar */}
      <div className="border-b border-border px-4 py-2 flex gap-2">
        {[
          { id: "public", label: "Public Chat", emoji: "🌐" },
          { id: "dm", label: "Messages", emoji: "📨" },
          { id: "contacts", label: "Contacts", emoji: "👥" },
        ].map(t => (
          <button
            key={t.id}
            onClick={() => { setTab(t.id); if (t.id !== "dm") setActiveDM(null); }}
            className={cn(
              "flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium transition-all",
              tab === t.id ? "bg-accent text-accent-foreground shadow-sm" : "text-muted-foreground hover:bg-muted"
            )}
          >
            {t.emoji} {t.label}
          </button>
        ))}
      </div>

      {/* Public Chat */}
      {tab === "public" && (
        <div className="flex flex-1 overflow-hidden">
          <div className="w-28 sm:w-36 border-r border-border p-2 space-y-0.5 flex-shrink-0">
            <p className="text-[10px] font-bold text-muted-foreground uppercase px-2 mb-1.5">Channels</p>
            {CHANNELS.map(ch => (
              <button
                key={ch.id}
                onClick={() => setChannel(ch.id)}
                className={cn(
                  "w-full flex items-center gap-1.5 px-2 py-1.5 rounded-lg text-xs font-medium transition-all",
                  channel === ch.id ? "bg-accent text-accent-foreground" : "text-muted-foreground hover:bg-muted/50"
                )}
              >
                <Hash className="h-3 w-3 flex-shrink-0" />{ch.label}
              </button>
            ))}
          </div>
          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="flex-1 overflow-y-auto p-4 space-y-3">
              {filteredMessages.map(msg => <ChatBubble key={msg.id} msg={msg} />)}
              {filteredMessages.length === 0 && (
                <div className="text-center py-12 text-muted-foreground">
                  <span className="text-3xl block mb-2">🤫</span>
                  <p className="text-sm">No messages yet in #{channel}</p>
                </div>
              )}
            </div>
            <div className="border-t border-border p-3">
              <div className="flex items-center gap-2">
                <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground flex-shrink-0">
                  <Image className="h-4 w-4" />
                </Button>
                <Input placeholder={`Message #${channel}...`} value={message} onChange={e => setMessage(e.target.value)} className="h-8 text-sm bg-muted/30" />
                <Button size="icon" className="h-8 w-8 bg-game-blue hover:bg-game-blue/90 text-white flex-shrink-0">
                  <Send className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* DMs */}
      {tab === "dm" && (
        <div className="flex flex-1 overflow-hidden">
          {/* Contact list — hidden on mobile when DM open */}
          <div className={cn("w-full sm:w-56 border-r border-border flex flex-col flex-shrink-0", activeDM ? "hidden sm:flex" : "flex")}>
            <div className="p-2 border-b border-border">
              <div className="relative">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                <Input placeholder="Search players..." value={search} onChange={e => setSearch(e.target.value)} className="h-8 pl-8 text-xs" />
              </div>
            </div>
            <div className="flex-1 overflow-y-auto py-1">
              {filteredContacts.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <span className="text-2xl block mb-1">🔍</span>
                  <p className="text-xs">No players found</p>
                </div>
              ) : filteredContacts.map(c => (
                <button
                  key={c.id}
                  onClick={() => setActiveDM(c.id)}
                  className={cn(
                    "w-full flex items-center gap-2.5 px-3 py-2.5 text-left transition-all hover:bg-muted/50",
                    activeDM === c.id ? "bg-accent" : ""
                  )}
                >
                  <div className="relative flex-shrink-0">
                    <span className="text-xl">{c.avatar}</span>
                    <span className={cn("absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-card", c.online ? "bg-game-green" : "bg-muted-foreground")} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <p className="text-xs font-semibold text-foreground truncate">{c.name}</p>
                      <span className="text-[9px] text-muted-foreground flex-shrink-0 ml-1">{c.time}</span>
                    </div>
                    <p className="text-[10px] text-muted-foreground truncate">{c.lastMsg}</p>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Conversation */}
          {activeDM ? (
            <div className="flex-1 flex flex-col overflow-hidden">
              <div className="flex items-center gap-2 px-3 py-2 border-b border-border bg-card/80">
                <Button variant="ghost" size="icon" className="sm:hidden h-7 w-7" onClick={() => setActiveDM(null)}>
                  <ArrowLeft className="h-4 w-4" />
                </Button>
                <span className="text-xl">{currentContact?.avatar}</span>
                <div>
                  <p className="text-sm font-bold text-foreground">{currentContact?.name}</p>
                  <p className="text-[10px] text-muted-foreground flex items-center gap-1">
                    <Circle className={cn("h-2 w-2 fill-current", currentContact?.online ? "text-game-green" : "text-muted-foreground")} />
                    {currentContact?.online ? "Online" : "Offline"} · {currentContact?.role}
                  </p>
                </div>
              </div>
              <div className="flex-1 overflow-y-auto p-3 space-y-3">
                {currentDMs.map(msg => <ChatBubble key={msg.id} msg={msg} />)}
              </div>
              <div className="border-t border-border p-3">
                <div className="flex items-center gap-2">
                  <Input
                    placeholder={`Message ${currentContact?.name}...`}
                    value={dmMessage}
                    onChange={e => setDmMessage(e.target.value)}
                    onKeyDown={e => e.key === "Enter" && sendDM()}
                    className="h-9 text-sm"
                  />
                  <Button size="icon" className="h-9 w-9 bg-game-blue hover:bg-game-blue/90 text-white flex-shrink-0" onClick={sendDM}>
                    <Send className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </div>
          ) : (
            <div className="hidden sm:flex flex-1 items-center justify-center text-center p-8 text-muted-foreground">
              <div>
                <span className="text-4xl block mb-3">📮</span>
                <p className="text-sm font-medium">Select a conversation</p>
                <p className="text-xs mt-1">Or search for a player to message</p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Contacts */}
      {tab === "contacts" && (
        <div className="flex-1 overflow-y-auto p-4">
          <div className="relative mb-3">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input placeholder="Search contacts..." className="pl-9 h-9" />
          </div>
          <div className="space-y-2 max-w-lg mx-auto">
            {DM_CONTACTS.map((c, i) => (
              <motion.div key={c.id} initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.05 }}>
                <Card className="p-3 border-border bg-card hover:shadow-sm transition-all">
                  <div className="flex items-center gap-3">
                    <div className="relative">
                      <span className="text-2xl">{c.avatar}</span>
                      <span className={cn("absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-card", c.online ? "bg-game-green" : "bg-muted")} />
                    </div>
                    <div className="flex-1">
                      <p className="font-semibold text-sm text-foreground">{c.name}</p>
                      <p className="text-xs text-muted-foreground">{c.role}</p>
                    </div>
                    <Button
                      size="sm"
                      className="bg-game-blue hover:bg-game-blue/90 text-white text-xs rounded-full px-3"
                      onClick={() => { setTab("dm"); setActiveDM(c.id); }}
                    >
                      Message
                    </Button>
                  </div>
                </Card>
              </motion.div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}