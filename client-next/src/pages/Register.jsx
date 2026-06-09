const db = globalThis.__B44_DB__ || { auth:{ isAuthenticated: async()=>false, me: async()=>null }, entities:new Proxy({}, { get:()=>({ filter:async()=>[], get:async()=>null, create:async()=>({}), update:async()=>({}), delete:async()=>({}) }) }), integrations:{ Core:{ UploadFile:async()=>({ file_url:'' }) } } };

import React, { useState } from "react";
import { Link } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Mail, Lock, User, Building2, Key, CheckCircle2 } from "lucide-react";

export default function Register() {
  const [step, setStep] = useState("form"); // "form" | "otp"
  const [form, setForm] = useState({ displayName: "", email: "", password: "", confirmPassword: "", companyName: "", inviteCode: "" });
  const [otpCode, setOtpCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const set = (key) => (e) => setForm((f) => ({ ...f, [key]: e.target.value }));

  const handleRegister = async (e) => {
    e.preventDefault();
    setError("");
    if (form.password !== form.confirmPassword) return setError("Passwords do not match.");
    if (form.password.length < 6) return setError("Password must be at least 6 characters.");
    setLoading(true);
    try {
      await db.auth.register({ email: form.email, password: form.password });
      setStep("otp");
    } catch (err) {
      setError(err.message || "Registration failed. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleVerify = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await db.auth.verifyOtp({ email: form.email, otpCode });
      db.auth.setToken(res.access_token);
      window.location.href = "/";
    } catch (err) {
      setError(err.message || "Invalid code. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        {/* Game Logo */}
        <div className="text-center mb-6">
          <div className="text-5xl mb-2">🏘️</div>
          <h1 className="text-2xl font-black text-primary font-heading">Harbor Town</h1>
          <p className="text-muted-foreground text-sm mt-1">Build your cozy food empire.</p>
          <div className="flex justify-center gap-2 mt-2 text-xl opacity-60">
            <span>🌾</span><span>🥛</span><span>🍞</span><span>☕</span><span>💰</span>
          </div>
        </div>

        {step === "form" ? (
          <div className="bg-card rounded-2xl border border-border shadow-lg p-6">
            <h2 className="text-xl font-bold text-foreground mb-1">Join the Beta 🎉</h2>
            <p className="text-xs text-muted-foreground mb-5">Create your harbor empire today.</p>

            {error && (
              <div className="mb-4 p-3 rounded-xl bg-destructive/10 text-destructive text-sm border border-destructive/20">{error}</div>
            )}

            <form onSubmit={handleRegister} className="space-y-3">
              <div>
                <Label className="text-xs font-semibold">Display Name</Label>
                <div className="relative mt-1">
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input placeholder="Captain Mochi" value={form.displayName} onChange={set("displayName")} className="pl-9 h-10" required />
                </div>
              </div>
              <div>
                <Label className="text-xs font-semibold">Email</Label>
                <div className="relative mt-1">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input type="email" placeholder="captain@harbor.town" value={form.email} onChange={set("email")} className="pl-9 h-10" required />
                </div>
              </div>
              <div>
                <Label className="text-xs font-semibold">Company Name</Label>
                <div className="relative mt-1">
                  <Building2 className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input placeholder="Mochi Foods Co." value={form.companyName} onChange={set("companyName")} className="pl-9 h-10" required />
                </div>
              </div>
              <div>
                <Label className="text-xs font-semibold">Password</Label>
                <div className="relative mt-1">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input type="password" placeholder="••••••••" value={form.password} onChange={set("password")} className="pl-9 h-10" required />
                </div>
              </div>
              <div>
                <Label className="text-xs font-semibold">Confirm Password</Label>
                <div className="relative mt-1">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input type="password" placeholder="••••••••" value={form.confirmPassword} onChange={set("confirmPassword")} className="pl-9 h-10" required />
                </div>
              </div>
              <div>
                <Label className="text-xs font-semibold">Invitation Code <span className="text-muted-foreground font-normal">(optional)</span></Label>
                <div className="relative mt-1">
                  <Key className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input placeholder="e.g. HARBOR-2024" value={form.inviteCode} onChange={set("inviteCode")} className="pl-9 h-10" />
                </div>
                <p className="text-[10px] text-muted-foreground mt-1">Enter your invitation code to join the closed test.</p>
              </div>
              <Button type="submit" className="w-full h-11 font-bold rounded-xl mt-1" disabled={loading}>
                {loading ? <><Loader2 className="w-4 h-4 mr-2 animate-spin" /> Creating account...</> : "🚢 Start My Empire!"}
              </Button>
            </form>
          </div>
        ) : (
          <div className="bg-card rounded-2xl border border-border shadow-lg p-6 text-center">
            <div className="text-4xl mb-3">📬</div>
            <h2 className="text-xl font-bold text-foreground mb-1">Check Your Email</h2>
            <p className="text-xs text-muted-foreground mb-5">
              We sent a verification code to <span className="font-semibold text-foreground">{form.email}</span>
            </p>

            {error && (
              <div className="mb-4 p-3 rounded-xl bg-destructive/10 text-destructive text-sm border border-destructive/20 text-left">{error}</div>
            )}

            <form onSubmit={handleVerify} className="space-y-4">
              <div>
                <Label className="text-xs font-semibold">Verification Code</Label>
                <Input
                  placeholder="Enter 6-digit code"
                  value={otpCode}
                  onChange={(e) => setOtpCode(e.target.value)}
                  className="h-12 text-center text-xl font-mono tracking-widest mt-1"
                  maxLength={6}
                  required
                />
              </div>
              <Button type="submit" className="w-full h-11 font-bold rounded-xl" disabled={loading}>
                {loading ? <><Loader2 className="w-4 h-4 mr-2 animate-spin" /> Verifying...</> : <><CheckCircle2 className="h-4 w-4 mr-2" /> Verify & Enter</>}
              </Button>
            </form>

            <button
              className="mt-3 text-xs text-primary hover:underline"
              onClick={async () => { await db.auth.resendOtp(form.email); }}
            >
              Resend code
            </button>
          </div>
        )}

        <p className="text-center text-sm text-muted-foreground mt-4">
          Already have an account?{" "}
          <Link to="/login" className="text-primary font-bold hover:underline">Log in</Link>
        </p>
      </div>
    </div>
  );
}