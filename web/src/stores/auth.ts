import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User, Membership } from "@/types/api";

interface AuthStore {
  user: User | null;
  memberships: Membership[];
  activeMembership: Membership | null;
  accessToken: string | null;
  refreshToken: string | null;
  setAuth: (user: User, accessToken: string, refreshToken: string) => void;
  setMemberships: (memberships: Membership[]) => void;
  setActiveMembership: (membership: Membership) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      user: null,
      memberships: [],
      activeMembership: null,
      accessToken: null,
      refreshToken: null,
      setAuth: (user, accessToken, refreshToken) =>
        set({ user, accessToken, refreshToken }),
      setMemberships: (memberships) => set({ memberships }),
      setActiveMembership: (membership) => {
        localStorage.setItem("company_id", membership.CompanyID);
        set({ activeMembership: membership });
      },
      clear: () => {
        localStorage.removeItem("company_id");
        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");
        set({
          user: null,
          memberships: [],
          activeMembership: null,
          accessToken: null,
          refreshToken: null,
        });
      },
    }),
    {
      name: "auth",
    }
  )
);