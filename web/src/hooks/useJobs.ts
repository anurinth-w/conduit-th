import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { Job } from "@/types/api";

export function useJobs() {
  const { activeMembership } = useAuthStore();
  const companyId = activeMembership?.CompanyID;

  return useQuery<Job[]>({
    queryKey: ["jobs", companyId],
    queryFn: async () => {
      const { data } = await api.get(`/v1/companies/${companyId}/jobs`);
      return data ?? [];
    },
    enabled: !!companyId,
  });
}
