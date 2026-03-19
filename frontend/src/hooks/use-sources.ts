import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getSources, createSource, toggleSource, getCollectorJobs, createCollectorJob, startCollectorJob } from "@/lib/api/sources";
import { mockSources, mockCollectorJobs } from "@/lib/mock-data";
import { ENABLE_MOCKS } from "@/lib/mock-config";

export const useSources = () =>
  useQuery({
    queryKey: ["sources"],
    queryFn: async () => {
      if (ENABLE_MOCKS) return mockSources;
      return await getSources();
    },
  });

export const useCreateSource = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ data }: { data: { type: string; name: string; url: string } }) => createSource(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sources"] }),
  });
};

export const useToggleSource = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => toggleSource(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sources"] }),
  });
};

export const useCollectorJobs = (sourceId?: string) =>
  useQuery({
    queryKey: ["collectorJobs", sourceId],
    queryFn: async () => {
      if (ENABLE_MOCKS) return mockCollectorJobs;
      return await getCollectorJobs(sourceId);
    },
  });

export const useStartCollectorJob = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (sourceId: string) => {
      const job = await createCollectorJob(sourceId);
      return startCollectorJob(job.id);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["collectorJobs"] }),
  });
};
