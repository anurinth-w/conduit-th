import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { JobPhoto } from "@/types/api";

export function usePhotos(jobId: string) {
  return useQuery<JobPhoto[]>({
    queryKey: ["photos", jobId],
    queryFn: async () => {
      const { data } = await api.get(`/v1/jobs/${jobId}/photos`);
      return data ?? [];
    },
    enabled: !!jobId,
  });
}

export function useTogglePhoto(jobId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ photoId, selected }: { photoId: string; selected: boolean }) => {
      await api.patch(`/v1/photos/${photoId}/select`, { is_selected: selected });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["photos", jobId] });
    },
  });
}
