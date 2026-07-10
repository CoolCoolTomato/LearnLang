import { http } from "./request"
import type { DeveloperDashboard, DeveloperPage, DeveloperResource } from "@/types/developer"

const basePath = "/dev"

export function getDeveloperDashboard() {
  return http.get<DeveloperDashboard>(`${basePath}/dashboard`)
}

export function listDeveloperRecords(resource: DeveloperResource, page: number, size: number) {
  return http.get<DeveloperPage>(`${basePath}/${resource}`, { params: { page, size } })
}

export function createDeveloperRecord(resource: DeveloperResource, values: Record<string, unknown>) {
  return http.post<Record<string, unknown>>(`${basePath}/${resource}`, values)
}

export function updateDeveloperRecord(resource: DeveloperResource, id: number, values: Record<string, unknown>) {
  return http.put<Record<string, unknown>>(`${basePath}/${resource}/${id}`, values)
}

export function deleteDeveloperRecord(resource: DeveloperResource, id: number) {
  return http.delete(`${basePath}/${resource}/${id}`)
}

export function deleteDeveloperRecords(resource: DeveloperResource, ids: number[]) {
  return http.delete<{ deleted: number }>(`${basePath}/${resource}`, { body: JSON.stringify({ ids }) })
}
