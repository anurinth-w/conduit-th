import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { Member } from "@/types/api";

export function useMembers() {
  const { activeMembership } = useAuthStore();
  const companyId = activeMembership?.CompanyID;

  return useQuery<Member[]>({
    queryKey: ["members", companyId],
    queryFn: async () => {
      const { data } = await api.get(`/v1/companies/${companyId}/members`);
      return data ?? [];
    },
    enabled: !!companyId,
  });
}
