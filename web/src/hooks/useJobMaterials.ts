import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { JobMaterial } from "@/types/api";

export function useJobMaterials(jobId: string) {
  return useQuery<JobMaterial[]>({
    queryKey: ["job-materials", jobId],
    queryFn: async () => {
      const { data } = await api.get(`/v1/jobs/${jobId}/materials`);
      return data ?? [];
    },
    enabled: !!jobId,
  });
}

export function useAddMaterial(jobId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      material_id?: string;
      code?: string;
      name: string;
      unit: string;
      quantity: number;
      unit_price?: number;
      labor_cost?: number;
    }) => {
      const { data } = await api.post(`/v1/jobs/${jobId}/materials`, payload);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["job-materials", jobId] });
    },
  });
}

export function useDeleteMaterial(jobId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (materialId: string) => {
      await api.delete(`/v1/jobs/${jobId}/materials/${materialId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["job-materials", jobId] });
    },
  });
}
