import api from "@/lib/api";

import type { components } from "./types";

export type EventType = string;

export type AppEvent = components["schemas"]["prod-pobeda-2026_internal_controller_http_v1_dto.EventResponse"];

export const getEvents = (type?: string, limit = 50, offset = 0) =>
  api.get<{ data: { items: AppEvent[] } }>("/events", { params: { type, limit, offset } }).then((r) => r.data.data?.items ?? []);
