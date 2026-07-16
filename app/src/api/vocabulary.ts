import { http } from "./request"
import type {
  VocabularyClearResult,
  VocabularyImportResult,
  VocabularyPage,
} from "@/types/vocabulary"

const basePath = "/vocabularies"

export function getVocabulary(page = 1, pageSize = 20) {
  return http.get<VocabularyPage>(basePath, {
    params: { page, page_size: pageSize },
  })
}

export function importVocabulary(data: unknown) {
  return http.post<VocabularyImportResult>(`${basePath}/import`, data, {
    timeout: 120_000,
  })
}

export function clearVocabulary() {
  return http.delete<VocabularyClearResult>(basePath)
}
