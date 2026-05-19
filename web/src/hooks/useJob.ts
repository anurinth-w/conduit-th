import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { Job } from "@/types/api";

export function useJob(id: string) {
  return useQuery<Job>({
    queryKey: ["job", id],
    queryFn: async () => {
      const { data } = await api.get(`/v1/jobs/${id}`);
      return data;
    },
    enabled: !!id,
  });
}
