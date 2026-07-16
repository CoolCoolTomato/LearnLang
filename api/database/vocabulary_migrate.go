package database

import "gorm.io/gorm"

func migrateLegacyVocabularySchema() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		legacyEntries, err := databaseColumnExists(tx, "vocabulary_entries", "native_text")
		if err != nil {
			return err
		}
		legacyVocabularies, err := databaseColumnExists(tx, "vocabularies", "source")
		if err != nil {
			return err
		}
		if !legacyEntries && !legacyVocabularies {
			return nil
		}

		if err := tx.Exec(`
			WITH selected AS (
				SELECT DISTINCT ON (v.user_id) v.id
				FROM vocabularies v
				WHERE NOT EXISTS (
					SELECT 1
					FROM vocabularies current_default
					WHERE current_default.user_id = v.user_id
						AND current_default.is_default = TRUE
				)
				ORDER BY v.user_id, v.id
			)
			UPDATE vocabularies
			SET is_default = TRUE
			WHERE id IN (SELECT id FROM selected)
		`).Error; err != nil {
			return err
		}

		if legacyEntries {
			if err := tx.Exec(`
				INSERT INTO vocabulary_meanings (
					entry_id,
					native_text,
					native_language,
					part_of_speech,
					sort_order,
					created_at,
					updated_at
				)
				SELECT
					entry.id,
					entry.native_text,
					entry.native_language,
					COALESCE(entry.part_of_speech, ''),
					0,
					COALESCE(entry.created_at, NOW()),
					COALESCE(entry.updated_at, NOW())
				FROM vocabulary_entries entry
				WHERE BTRIM(entry.native_text) <> ''
					AND NOT EXISTS (
						SELECT 1
						FROM vocabulary_meanings meaning
						WHERE meaning.entry_id = entry.id
							AND LOWER(BTRIM(meaning.native_text)) = LOWER(BTRIM(entry.native_text))
							AND LOWER(BTRIM(meaning.native_language)) = LOWER(BTRIM(entry.native_language))
							AND LOWER(BTRIM(COALESCE(meaning.part_of_speech, ''))) = LOWER(BTRIM(COALESCE(entry.part_of_speech, '')))
					)
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				INSERT INTO vocabulary_pronunciations (
					entry_id,
					pronunciation,
					pronunciation_type,
					region,
					sort_order,
					created_at,
					updated_at
				)
				SELECT
					entry.id,
					entry.pronunciation,
					'ipa',
					'',
					0,
					COALESCE(entry.created_at, NOW()),
					COALESCE(entry.updated_at, NOW())
				FROM vocabulary_entries entry
				WHERE BTRIM(COALESCE(entry.pronunciation, '')) <> ''
				ON CONFLICT (entry_id, pronunciation_type, region) DO NOTHING
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				INSERT INTO vocabulary_examples (
					entry_id,
					target_text,
					native_text,
					source,
					sort_order,
					created_at,
					updated_at
				)
				SELECT
					entry.id,
					entry.target_example,
					COALESCE(entry.native_example, ''),
					entry.source,
					0,
					COALESCE(entry.created_at, NOW()),
					COALESCE(entry.updated_at, NOW())
				FROM vocabulary_entries entry
				WHERE BTRIM(COALESCE(entry.target_example, '')) <> ''
					AND NOT EXISTS (
						SELECT 1
						FROM vocabulary_examples example
						WHERE example.entry_id = entry.id
							AND BTRIM(example.target_text) = BTRIM(entry.target_example)
							AND BTRIM(COALESCE(example.native_text, '')) = BTRIM(COALESCE(entry.native_example, ''))
					)
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				DROP INDEX IF EXISTS idx_vocabulary_entries_user_encountered;
				DROP INDEX IF EXISTS idx_vocabulary_entry_identity;
				ALTER TABLE vocabulary_entries
					DROP COLUMN IF EXISTS user_id,
					DROP COLUMN IF EXISTS native_text,
					DROP COLUMN IF EXISTS native_language,
					DROP COLUMN IF EXISTS pronunciation,
					DROP COLUMN IF EXISTS part_of_speech,
					DROP COLUMN IF EXISTS target_example,
					DROP COLUMN IF EXISTS native_example;
			`).Error; err != nil {
				return err
			}
		}

		if legacyVocabularies {
			if err := tx.Exec(`
				ALTER TABLE vocabularies
					DROP COLUMN IF EXISTS source;
			`).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func databaseColumnExists(tx *gorm.DB, table, column string) (bool, error) {
	var count int64
	err := tx.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = CURRENT_SCHEMA()
			AND table_name = ?
			AND column_name = ?
	`, table, column).Scan(&count).Error
	return count > 0, err
}
