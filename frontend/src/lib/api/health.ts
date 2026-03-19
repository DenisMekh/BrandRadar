import api from "@/lib/api";

import type { components } from "./types";

export type HealthStatus = components["schemas"]["prod-pobeda-2026_internal_controller_http_v1_dto.HealthResponse"];

export const getHealth = () =>
  api.get<{ data: HealthStatus }>("/health").then((r) => r.data.data);
