import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { PLAYER } from "@/lib/gameData";

export default function SettingsPage() {
  return (
    <div className="p-4 max-w-2xl mx-auto space-y-4">
      <h1 className="text-xl font-bold text-foreground">⚙️ Settings</h1>

      {/* Profile */}
      <Card className="p-4 bg-card border-border">
        <p className="text-xs font-semibold text-muted-foreground mb-3">Profile</p>
        <div className="flex items-center gap-4 mb-4">
          <div className="w-16 h-16 rounded-2xl bg-accent flex items-center justify-center text-4xl">{PLAYER.avatar}</div>
          <div>
            <p className="font-bold text-foreground">{PLAYER.name}</p>
            <p className="text-xs text-muted-foreground">Level {PLAYER.level}</p>
          </div>
        </div>
        <div className="space-y-3">
          <div>
            <Label className="text-xs">Display Name</Label>
            <Input defaultValue={PLAYER.name} className="h-8 text-sm mt-1" />
          </div>
          <Button size="sm" className="bg-primary hover:bg-primary/90 text-primary-foreground text-xs">Save Changes</Button>
        </div>
      </Card>

      {/* Preferences */}
      <Card className="p-4 bg-card border-border">
        <p className="text-xs font-semibold text-muted-foreground mb-3">Preferences</p>
        <div className="space-y-4">
          {[
            { label: "Sound Effects", desc: "Play sounds for actions", defaultChecked: true },
            { label: "Music", desc: "Background music", defaultChecked: false },
            { label: "Notifications", desc: "Push notifications", defaultChecked: true },
            { label: "Price Alerts", desc: "Alert when prices change significantly", defaultChecked: true },
          ].map((s) => (
            <div key={s.label} className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-foreground">{s.label}</p>
                <p className="text-xs text-muted-foreground">{s.desc}</p>
              </div>
              <Switch defaultChecked={s.defaultChecked} />
            </div>
          ))}
        </div>
      </Card>

      {/* Danger Zone */}
      <Card className="p-4 bg-card border-game-red/20">
        <p className="text-xs font-semibold text-game-red mb-3">Danger Zone</p>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-foreground">Reset Game</p>
            <p className="text-xs text-muted-foreground">This will delete all your progress</p>
          </div>
          <Button variant="outline" size="sm" className="text-game-red border-game-red/30 hover:bg-game-red/10 text-xs">
            Reset
          </Button>
        </div>
      </Card>
    </div>
  );
}