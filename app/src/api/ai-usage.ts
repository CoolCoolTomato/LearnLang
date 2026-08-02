import { http } from "./request"
import type { AIUsagePage, AIUsageSummary } from "@/types/ai-usage"

export function listAIUsage(page = 1, pageSize = 20, operation = "") {
  return http.get<AIUsagePage>("/usage/events", {
    params: { page, page_size: pageSize, ...(operation ? { operation } : {}) },
  })
}

export function summarizeAIUsage(operation = "") {
  return http.get<AIUsageSummary[]>("/usage/summary", {
    params: operation ? { operation } : undefined,
  })
}
