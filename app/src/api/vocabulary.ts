import { http } from "./request"
import type {
  VocabularyClearResult,
  VocabularyImportResult,
  Vocabulary,
  VocabularyInput,
  VocabularyPage,
  VocabularyLookupResult,
  VocabularySummary,
  VocabularyUpdateInput,
} from "@/types/vocabulary"

const basePath = "/vocabularies"

export function listVocabularies() {
  return http.get<{ data: VocabularySummary[] }>(basePath)
}

export function createVocabulary(data: VocabularyInput) {
  return http.post<Vocabulary>(basePath, data)
}

export function updateVocabulary(
  vocabularyId: number,
  data: VocabularyUpdateInput
) {
  return http.put<Vocabulary>(`${basePath}/${vocabularyId}`, data)
}

export function setDefaultVocabulary(vocabularyId: number) {
  return http.put<void>(`${basePath}/${vocabularyId}/default`)
}

export function deleteVocabulary(vocabularyId: number) {
  return http.delete<void>(`${basePath}/${vocabularyId}`)
}

export function getVocabularyEntries(
  vocabularyId: number,
  page = 1,
  pageSize = 20,
  query = ""
) {
  const normalizedQuery = query.trim()
  return http.get<VocabularyPage>(`${basePath}/${vocabularyId}/entries`, {
    params: {
      page,
      page_size: pageSize,
      ...(normalizedQuery ? { query: normalizedQuery } : {}),
    },
  })
}

export function importVocabulary(vocabularyId: number, data: unknown) {
  return http.post<VocabularyImportResult>(
    `${basePath}/${vocabularyId}/import`,
    data,
    { timeout: 120_000 }
  )
}

export function clearVocabulary(vocabularyId: number) {
  return http.delete<VocabularyClearResult>(
    `${basePath}/${vocabularyId}/entries`
  )
}

export function lookupMessageVocabulary(messageId: number) {
  return http.post<VocabularyLookupResult>(`${basePath}/lookup`, {
    message_id: messageId,
  })
}
