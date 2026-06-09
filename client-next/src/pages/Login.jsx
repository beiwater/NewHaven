const db = globalThis.__B44_DB__ || { auth:{ isAuthenticated: async()=>false, me: async()=>null }, entities:new Proxy({}, { get:()=>({ filter:async()=>[], get:async()=>null, create:async()=>({}), update:async()=>({}), delete:async()=>({}) }) }), integrations:{ Core:{ UploadFile:async()=>({ file_url:'' }) } } };

import React, { useState } from "react";
import { Link } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Mail, Lock } from "lucide-react";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await db.auth.loginViaEmailPassword(email, password);
      window.location.href = "/";
    } catch (err) {
      setError(err.message || "Invalid email or password");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        {/* Game Logo */}
        <div className="text-center mb-8">
          <div className="text-6xl mb-3 animate-bounce-soft">🏘️</div>
          <h1 className="text-3xl font-black text-primary font-heading">Harbor Town</h1>
          <p className="text-muted-foreground text-sm mt-1">Build your cozy food empire.</p>
          <div className="flex justify-center gap-2 mt-3 text-2xl opacity-60">
            <span>🌾</span><span>🥛</span><span>🍞</span><span>☕</span><span>💰</span>
          </div>
        </div>

        {/* Card */}
        <div className="bg-card rounded-2xl border border-border shadow-lg p-6">
          <h2 className="text-xl font-bold text-foreground mb-1">Welcome back! 👋</h2>
          <p className="text-xs text-muted-foreground mb-5">Log in to continue your journey.</p>

          {error && (
            <div className="mb-4 p-3 rounded-xl bg-destructive/10 text-destructive text-sm border border-destructive/20">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <Label className="text-xs font-semibold">Email</Label>
              <div className="relative mt-1">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  type="email"
                  autoComplete="email"
                  placeholder="captain@harbor.town"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="pl-9 h-11"
                  required
                />
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between mb-1">
                <Label className="text-xs font-semibold">Password</Label>
                <Link to="/forgot-password" className="text-[11px] text-primary hover:underline">Forgot password?</Link>
              </div>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  type="password"
                  autoComplete="current-password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="pl-9 h-11"
                  required
                />
              </div>
            </div>
            <Button type="submit" className="w-full h-11 font-bold rounded-xl" disabled={loading}>
              {loading ? <><Loader2 className="w-4 h-4 mr-2 animate-spin" /> Logging in...</> : "⚓ Set Sail!"}
            </Button>
          </form>

          <div className="relative my-4">
            <div className="absolute inset-0 flex items-center"><div className="w-full border-t border-border" /></div>
            <div className="relative flex justify-center text-xs"><span className="bg-card px-3 text-muted-foreground">or</span></div>
          </div>

          <Button variant="outline" className="w-full h-11 rounded-xl" onClick={() => db.auth.loginWithProvider("google", "/")}>
            <span className="mr-2">🌐</span> Continue with Google
          </Button>
        </div>

        <p className="text-center text-sm text-muted-foreground mt-4">
          New to Harbor Town?{" "}
          <Link to="/register" className="text-primary font-bold hover:underline">Join the beta</Link>
        </p>
      </div>
    </div>
  );
}