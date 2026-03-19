import { useQuery } from "@tanstack/react-query";
import { getEvents } from "@/lib/api/events";
import { mockEvents } from "@/lib/mock-data";
import { ENABLE_MOCKS } from "@/lib/mock-config";

export const useEvents = (type?: string, limit = 50, offset = 0) =>
  useQuery({
    queryKey: ["events", type, limit, offset],
    queryFn: async () => {
      if (ENABLE_MOCKS) return [...mockEvents];
      return await getEvents(type, limit, offset);
    },
  });
