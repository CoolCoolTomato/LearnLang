export type VocabularyEntryType = "word" | "phrase"
export type VocabularySource = "import" | "chat" | "manual" | "system"

export interface Vocabulary {
  id: number
  user_id: number
  name: string
  target_language: string
  native_language: string
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface VocabularySummary {
  id: number
  name: string
  target_language: string
  native_language: string
  is_default: boolean
  entry_count: number
}

export interface VocabularyInput {
  name: string
  target_language?: string
  native_language?: string
  is_default?: boolean
}

export interface VocabularyUpdateInput {
  name?: string
  target_language?: string
  native_language?: string
  is_default?: boolean
}

export interface VocabularyPronunciation {
  id: number
  entry_id: number
  pronunciation: string
  pronunciation_type: string
  region: string
  audio_url?: string
  sort_order: number
}

export interface VocabularyMeaning {
  id: number
  entry_id: number
  native_text: string
  native_language: string
  part_of_speech?: string
  sort_order: number
}

export interface VocabularyExample {
  id: number
  entry_id: number
  meaning_id?: number
  target_text: string
  native_text?: string
  source: VocabularySource
  sort_order: number
}

export interface VocabularyRelatedEntry {
  id: number
  target_text: string
  target_language: string
  entry_type: VocabularyEntryType
  meanings?: VocabularyMeaning[]
}

export interface VocabularyEntryRelation {
  id: number
  entry_id: number
  related_entry_id: number
  relation_type: string
  sort_order: number
  related_entry?: VocabularyRelatedEntry
}

export interface VocabularyEntry {
  id: number
  vocabulary_id: number
  target_text: string
  target_language: string
  entry_type: VocabularyEntryType
  tags?: string[] | null
  notes?: string
  source: VocabularySource
  encountered: boolean
  encounter_count: number
  first_encountered_at?: string
  last_encountered_at?: string
  pronunciations?: VocabularyPronunciation[]
  meanings?: VocabularyMeaning[]
  examples?: VocabularyExample[]
  relations?: VocabularyEntryRelation[]
  created_at: string
  updated_at: string
}

export interface VocabularyPage {
  vocabulary: Vocabulary
  data: VocabularyEntry[]
  total: number
  page: number
  page_size: number
}

export interface VocabularyImportResult {
  vocabulary: Vocabulary
  entries_created: number
  entries_updated: number
  meanings_created: number
  pronunciations_created: number
  examples_created: number
  relations_created: number
}

export interface VocabularyClearResult {
  deleted: number
}
