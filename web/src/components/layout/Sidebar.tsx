"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { cn } from "@/lib/utils";
import { Briefcase, Package, Users, LogOut } from "lucide-react";

const navItems = [
  { href: "/jobs", label: "รายการงาน", icon: Briefcase },
  { href: "/materials", label: "วัสดุ", icon: Package },
  { href: "/users", label: "ผู้ใช้งาน", icon: Users, roles: ["admin", "manager"] },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user, activeMembership, clear } = useAuthStore();

  function handleLogout() {
    clear();
    window.location.href = "/login";
  }

  return (
    <aside className="w-60 min-h-screen bg-card border-r flex flex-col">
      <div className="px-6 py-5 border-b">
        <h1 className="font-semibold text-lg">Conduit-TH</h1>
        <p className="text-xs text-muted-foreground mt-0.5 truncate">
          {activeMembership?.CompanyName ?? "—"}
        </p>
      </div>

      <nav className="flex-1 px-3 py-4 space-y-1">
        {navItems.map(({ href, label, icon: Icon, roles }) => {
          const role = activeMembership?.Role;
          if (roles && role && !roles.includes(role)) return null;

          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors",
                pathname.startsWith(href)
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground"
              )}
            >
              <Icon className="w-4 h-4 shrink-0" />
              {label}
            </Link>
          );
        })}
      </nav>

      <div className="px-4 py-4 border-t">
        <p className="text-sm font-medium truncate">{user?.name}</p>
        <p className="text-xs text-muted-foreground truncate">{activeMembership?.Role}</p>
        <button
          onClick={handleLogout}
          className="mt-3 flex items-center gap-2 text-xs text-muted-foreground hover:text-foreground transition-colors"
        >
          <LogOut className="w-3.5 h-3.5" />
          ออกจากระบบ
        </button>
      </div>
    </aside>
  );
}
