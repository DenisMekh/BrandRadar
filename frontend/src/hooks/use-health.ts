import { useQuery } from "@tanstack/react-query";
import { getHealth } from "@/lib/api/health";
import { mockHealth } from "@/lib/mock-data";
import { ENABLE_MOCKS } from "@/lib/mock-config";

export const useHealth = () =>
  useQuery({
    queryKey: ["health"],
    queryFn: async () => {
      if (ENABLE_MOCKS) return mockHealth;
      return await getHealth();
    },
    refetchInterval: 30000,
  });
