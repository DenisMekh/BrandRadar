import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getMentions, updateMentionStatus, type MentionFilters } from "@/lib/api/mentions";

export const useMentions = (filters: MentionFilters) =>
  useQuery({
    queryKey: ["mentions", filters],
    queryFn: () => getMentions(filters),
  });

export const useUpdateMentionStatus = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => updateMentionStatus(id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["mentions"] }),
  });
};
