import axios from "axios";
import { setCache } from "@/lib/api-cache";

const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

const api = axios.create({
  baseURL: API_BASE_URL.startsWith('http') ? `${API_BASE_URL}/api/v1` : `${API_BASE_URL}/api/v1`.replace(/^\/+/, '/'),
  timeout: 10000,
  headers: { "Content-Type": "application/json" },
});

api.interceptors.response.use(
  (response) => {
    if (response.data?.error) {
      throw new Error(response.data.error);
    }
    // Cache successful GET responses
    if (response.config.method === "get" && response.config.url) {
      const key = response.config.url + (response.config.params ? JSON.stringify(response.config.params) : "");
      setCache(key, response.data);
    }
    return response;
  },
  (error) => {
    if (error.response?.data?.error) {
      throw new Error(error.response.data.error);
    }
    throw error;
  }
);

export default api;
