"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { Membership } from "@/types/api";

export default function LoginPage() {
  const router = useRouter();
  const { setAuth, setMemberships, setActiveMembership } = useAuthStore();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      const { data } = await api.post("/v1/auth/login", { email, password });

      const payload = JSON.parse(atob(data.access_token.split(".")[1]));

      localStorage.setItem("access_token", data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
      document.cookie = `access_token=${data.access_token}; path=/; max-age=900; SameSite=Lax`;

      setAuth(
        { uid: payload.uid, email: payload.email, name: payload.name },
        data.access_token,
        data.refresh_token
      );

      const { data: memberships } = await api.get<Membership[]>(
        `/v1/users/${payload.uid}/memberships`
      );
      setMemberships(memberships ?? []);

      if (!memberships || memberships.length === 0) {
        toast.error("ไม่พบบริษัทที่เชื่อมกับบัญชีนี้");
        return;
      }

      if (memberships.length === 1) {
        setActiveMembership(memberships[0]);
        router.push("/jobs");
      } else {
        router.push("/select-company");
      }
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data
          ?.error ?? "เข้าสู่ระบบไม่สำเร็จ";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Conduit-TH</CardTitle>
          <p className="text-sm text-muted-foreground">ระบบจัดการงานภาคสนาม</p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleLogin} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="email">อีเมล</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="password">รหัสผ่าน</Label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "กำลังเข้าสู่ระบบ..." : "เข้าสู่ระบบ"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
