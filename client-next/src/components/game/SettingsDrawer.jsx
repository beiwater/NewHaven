import { useState } from "react";
import { X, User, Building2, Globe, Volume2, VolumeX, LogOut, Trash2, ChevronRight, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { cn } from "@/lib/utils";
import { PLAYER } from "@/lib/gameData";
import { motion, AnimatePresence } from "framer-motion";

const SECTIONS = [
  { id: "account", label: "Account", icon: "👤" },
  { id: "company", label: "Company", icon: "🏢" },
  { id: "language", label: "Language", icon: "🌐" },
  { id: "audio", label: "Audio", icon: "🔊" },
  { id: "danger", label: "Delete Company", icon: "⚠️", danger: true },
];

const AVATARS = ["🧑🍳", "👨🌾", "👩💼", "🏴☠️", "🧙♂️", "👸", "🤴", "🧚", "🧜♀️", "🦊"];
const LANGUAGES = [
  { code: "en", label: "English", flag: "🇬🇧", available: true },
  { code: "zh", label: "中文 (Chinese)", flag: "🇨🇳", available: false },
  { code: "ja", label: "日本語 (Japanese)", flag: "🇯🇵", available: false },
  { code: "es", label: "Español (Spanish)", flag: "🇪🇸", available: false },
];

export default function SettingsDrawer({ open, onClose }) {
  const [section, setSection] = useState("account");
  const [saved, setSaved] = useState(false);
  const [displayName, setDisplayName] = useState(PLAYER.name);
  const [companyName, setCompanyName] = useState("Mochi Foods Co.");
  const [selectedAvatar, setSelectedAvatar] = useState(PLAYER.avatar);
  const [lang, setLang] = useState("en");
  const [clickVol, setClickVol] = useState([60]);
  const [musicVol, setMusicVol] = useState([40]);
  const [clickMuted, setClickMuted] = useState(false);
  const [musicMuted, setMusicMuted] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState("");

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const deleteReady = deleteConfirm === companyName;

  return (
    <AnimatePresence>
      {open && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/30 z-[60]"
            onClick={onClose}
          />
          {/* Drawer */}
          <motion.div
            initial={{ x: "100%" }}
            animate={{ x: 0 }}
            exit={{ x: "100%" }}
            transition={{ type: "spring", damping: 25, stiffness: 200 }}
            className="fixed top-0 right-0 bottom-0 w-full max-w-sm bg-card border-l border-border z-[70] flex flex-col shadow-2xl"
          >
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-primary">
              <h2 className="font-bold text-primary-foreground">⚙️ Settings</h2>
              <Button variant="ghost" size="icon" className="h-8 w-8 text-primary-foreground/70 hover:text-primary-foreground hover:bg-primary-foreground/10" onClick={onClose}>
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="flex flex-1 overflow-hidden">
              {/* Section Nav */}
              <div className="w-24 border-r border-border bg-muted/30 py-2 flex flex-col gap-1 flex-shrink-0">
                {SECTIONS.map((s) => (
                  <button
                    key={s.id}
                    onClick={() => setSection(s.id)}
                    className={cn(
                      "flex flex-col items-center gap-1 px-2 py-2.5 text-center text-[10px] font-medium rounded-lg mx-1 transition-all",
                      section === s.id
                        ? s.danger ? "bg-game-red/10 text-game-red" : "bg-accent text-accent-foreground"
                        : s.danger ? "text-game-red/70 hover:bg-game-red/5" : "text-muted-foreground hover:bg-muted"
                    )}
                  >
                    <span className="text-lg">{s.icon}</span>
                    <span className="leading-tight">{s.label}</span>
                  </button>
                ))}
              </div>

              {/* Content */}
              <div className="flex-1 overflow-y-auto p-4">
                <AnimatePresence mode="wait">
                  <motion.div key={section} initial={{ opacity: 0, x: 10 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -10 }} transition={{ duration: 0.15 }}>

                    {/* Account */}
                    {section === "account" && (
                      <div className="space-y-4">
                        <h3 className="font-bold text-sm text-foreground">Account Settings</h3>
                        <div className="space-y-3">
                          <div>
                            <Label className="text-xs">Display Name</Label>
                            <Input value={displayName} onChange={e => setDisplayName(e.target.value)} className="h-9 text-sm mt-1" />
                          </div>
                          <div>
                            <Label className="text-xs">Email</Label>
                            <Input defaultValue="captain@harbor.town" type="email" className="h-9 text-sm mt-1" />
                          </div>
                          <div>
                            <Label className="text-xs">New Password</Label>
                            <Input type="password" placeholder="Leave blank to keep current" className="h-9 text-sm mt-1" />
                          </div>
                          <div>
                            <Label className="text-xs">Confirm Password</Label>
                            <Input type="password" placeholder="Repeat new password" className="h-9 text-sm mt-1" />
                          </div>
                          <Button onClick={handleSave} className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm">
                            {saved ? <><CheckCircle2 className="h-4 w-4" /> Saved!</> : "Save Changes"}
                          </Button>
                        </div>
                      </div>
                    )}

                    {/* Company */}
                    {section === "company" && (
                      <div className="space-y-4">
                        <h3 className="font-bold text-sm text-foreground">Company Settings</h3>
                        <div className="space-y-3">
                          <div>
                            <Label className="text-xs">Company Name</Label>
                            <Input value={companyName} onChange={e => setCompanyName(e.target.value)} className="h-9 text-sm mt-1" />
                          </div>
                          <div>
                            <Label className="text-xs">Character Name</Label>
                            <Input value={displayName} onChange={e => setDisplayName(e.target.value)} className="h-9 text-sm mt-1" />
                          </div>
                          <div>
                            <Label className="text-xs">Role / Title</Label>
                            <Input defaultValue="Harbor Tycoon" className="h-9 text-sm mt-1" />
                          </div>
                          <div>
                            <Label className="text-xs mb-2 block">Choose Avatar</Label>
                            <div className="grid grid-cols-5 gap-1.5">
                              {AVATARS.map((av) => (
                                <button
                                  key={av}
                                  onClick={() => setSelectedAvatar(av)}
                                  className={cn(
                                    "aspect-square rounded-xl text-2xl flex items-center justify-center transition-all",
                                    selectedAvatar === av ? "bg-accent ring-2 ring-primary" : "bg-muted hover:bg-muted/70"
                                  )}
                                >{av}</button>
                              ))}
                            </div>
                          </div>
                          {/* Preview */}
                          <div className="rounded-xl bg-muted/50 p-3 flex items-center gap-3">
                            <span className="text-3xl">{selectedAvatar}</span>
                            <div>
                              <p className="font-bold text-sm text-foreground">{displayName}</p>
                              <p className="text-xs text-muted-foreground">{companyName}</p>
                              <p className="text-[10px] text-primary">Lv.{PLAYER.level} · Harbor Tycoon</p>
                            </div>
                          </div>
                          <Button onClick={handleSave} className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm">
                            {saved ? <><CheckCircle2 className="h-4 w-4" /> Saved!</> : "Save Changes"}
                          </Button>
                        </div>
                      </div>
                    )}

                    {/* Language */}
                    {section === "language" && (
                      <div className="space-y-4">
                        <h3 className="font-bold text-sm text-foreground">Language</h3>
                        <div className="space-y-1.5">
                          {LANGUAGES.map((l) => (
                            <button
                              key={l.code}
                              onClick={() => l.available && setLang(l.code)}
                              className={cn(
                                "w-full flex items-center gap-3 px-3 py-2.5 rounded-xl border text-left transition-all",
                                !l.available && "opacity-50 cursor-not-allowed",
                                lang === l.code && l.available ? "border-primary bg-accent" : "border-border bg-card hover:bg-muted/50"
                              )}
                            >
                              <span className="text-xl">{l.flag}</span>
                              <div className="flex-1">
                                <p className="text-sm font-medium text-foreground">{l.label}</p>
                                {!l.available && <p className="text-[10px] text-muted-foreground">Coming soon</p>}
                              </div>
                              {lang === l.code && l.available && <CheckCircle2 className="h-4 w-4 text-primary" />}
                            </button>
                          ))}
                        </div>
                        <p className="text-[11px] text-muted-foreground bg-muted/50 rounded-lg p-2.5">
                          🌏 More languages coming soon! The app is prepared for full translation support.
                        </p>
                      </div>
                    )}

                    {/* Audio */}
                    {section === "audio" && (
                      <div className="space-y-5">
                        <h3 className="font-bold text-sm text-foreground">Audio Settings</h3>

                        {/* Click Sound */}
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <Label className="text-xs font-semibold">Click Sound</Label>
                            <div className="flex items-center gap-1.5">
                              <Button
                                variant="ghost"
                                size="icon"
                                className={cn("h-7 w-7", clickMuted ? "text-game-red" : "text-muted-foreground")}
                                onClick={() => setClickMuted(!clickMuted)}
                              >
                                {clickMuted ? <VolumeX className="h-4 w-4" /> : <Volume2 className="h-4 w-4" />}
                              </Button>
                              <Button variant="outline" size="sm" className="text-[10px] h-6 px-2">
                                Test 🔊
                              </Button>
                            </div>
                          </div>
                          <div className="flex items-center gap-3">
                            <Slider
                              value={clickMuted ? [0] : clickVol}
                              onValueChange={setClickVol}
                              max={100} step={1}
                              className={cn("flex-1", clickMuted && "opacity-40")}
                              disabled={clickMuted}
                            />
                            <span className="text-xs font-mono text-muted-foreground w-8 text-right">{clickMuted ? "0%" : `${clickVol[0]}%`}</span>
                          </div>
                        </div>

                        {/* BGM */}
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <Label className="text-xs font-semibold">Background Music</Label>
                            <Button
                              variant="ghost"
                              size="icon"
                              className={cn("h-7 w-7", musicMuted ? "text-game-red" : "text-muted-foreground")}
                              onClick={() => setMusicMuted(!musicMuted)}
                            >
                              {musicMuted ? <VolumeX className="h-4 w-4" /> : <Volume2 className="h-4 w-4" />}
                            </Button>
                          </div>
                          <div className="flex items-center gap-3">
                            <Slider
                              value={musicMuted ? [0] : musicVol}
                              onValueChange={setMusicVol}
                              max={100} step={1}
                              className={cn("flex-1", musicMuted && "opacity-40")}
                              disabled={musicMuted}
                            />
                            <span className="text-xs font-mono text-muted-foreground w-8 text-right">{musicMuted ? "0%" : `${musicVol[0]}%`}</span>
                          </div>
                        </div>

                        <Button onClick={handleSave} className="w-full bg-game-blue hover:bg-game-blue/90 text-white gap-2" size="sm">
                          {saved ? <><CheckCircle2 className="h-4 w-4" /> Saved!</> : "Save Settings"}
                        </Button>
                      </div>
                    )}

                    {/* Danger */}
                    {section === "danger" && (
                      <div className="space-y-4">
                        <div className="flex items-center gap-2">
                          <span className="text-xl">⚠️</span>
                          <h3 className="font-bold text-sm text-game-red">Delete Company</h3>
                        </div>
                        <div className="bg-game-red/5 border border-game-red/20 rounded-xl p-3 space-y-1">
                          <p className="text-xs font-semibold text-game-red">This action cannot be undone!</p>
                          <p className="text-xs text-muted-foreground">All your buildings, resources, orders, and progress will be permanently deleted.</p>
                        </div>

                        {/* Logout first */}
                        <div className="border-t border-border pt-4">
                          <p className="text-xs font-semibold text-muted-foreground mb-2">Logout</p>
                          <Button variant="outline" className="w-full gap-2 text-xs" size="sm">
                            <LogOut className="h-3.5 w-3.5" /> Log Out of Game
                          </Button>
                        </div>

                        <div className="border-t border-game-red/20 pt-4 space-y-3">
                          <p className="text-xs text-foreground">
                            Type your company name <span className="font-bold text-game-red">'{companyName}'</span> to confirm deletion:
                          </p>
                          <Input
                            value={deleteConfirm}
                            onChange={e => setDeleteConfirm(e.target.value)}
                            placeholder={`Type "${companyName}" here`}
                            className="h-9 text-sm border-game-red/30 focus:border-game-red"
                          />
                          <Button
                            disabled={!deleteReady}
                            className={cn(
                              "w-full gap-2 text-xs",
                              deleteReady
                                ? "bg-game-red hover:bg-game-red/90 text-white"
                                : "bg-muted text-muted-foreground cursor-not-allowed"
                            )}
                            size="sm"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                            {deleteReady ? "Delete My Company Forever" : "Enter company name to unlock"}
                          </Button>
                        </div>
                      </div>
                    )}

                  </motion.div>
                </AnimatePresence>
              </div>
            </div>

            {/* Footer logout */}
            <div className="border-t border-border p-3">
              <Button variant="ghost" className="w-full gap-2 text-sm text-muted-foreground hover:text-game-red hover:bg-game-red/5" size="sm">
                <LogOut className="h-4 w-4" /> Log Out
              </Button>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}