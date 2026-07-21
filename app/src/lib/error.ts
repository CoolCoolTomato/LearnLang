import i18n from '@/i18n'
import { ApiError } from '@/api/request'

const exactApiErrorKeys: Record<string, string> = {
  'invalid credentials': 'errors.invalidCredentials',
  'email already exists': 'errors.emailExists',
  'phone already exists': 'errors.phoneExists',
  'email or phone is required': 'errors.contactRequired',
  'invalid email': 'errors.invalidEmail',
  'invalid phone': 'errors.invalidPhone',
  'translation text is required': 'errors.translationTextRequired',
  'vocabulary not found': 'errors.vocabularyNotFound',
  'message not found': 'errors.messageNotFound',
  'a vocabulary with this name already exists': 'errors.vocabularyNameConflict',
  'target language and native language are required': 'errors.vocabularyLanguagesRequired',
  'a default vocabulary is required': 'errors.defaultVocabularyRequired',
}

const prefixedApiErrorKeys: Array<[string, string]> = [
  ['translation text is too long', 'errors.translationTextTooLong'],
  ['invalid vocabulary import', 'errors.vocabularyInvalidImport'],
  ['invalid vocabulary input', 'errors.vocabularyInvalidInput'],
]

const statusErrorKeys: Record<number, string> = {
  403: 'errors.forbidden',
  404: 'errors.notFound',
  408: 'errors.requestTimeout',
  409: 'errors.conflict',
  429: 'errors.tooManyRequests',
}

function getApiErrorKey(error: ApiError): string | undefined {
  const message = error.message.trim().toLowerCase()
  const exactKey = exactApiErrorKeys[message]
  if (exactKey) return exactKey

  const prefixedKey = prefixedApiErrorKeys.find(([prefix]) =>
    message.startsWith(prefix)
  )?.[1]
  if (prefixedKey) return prefixedKey

  return error.status ? statusErrorKeys[error.status] : undefined
}

export function getErrorMessage(error: unknown, fallback: string): string {
  if (!(error instanceof ApiError)) return fallback

  const key = getApiErrorKey(error)
  return key ? i18n.t(key) : fallback
}
