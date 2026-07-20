import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import enUS from './locales/en-US.json'
import zhCN from './locales/zh-CN.json'

export const LANGUAGE_KEY = 'app_language'

const getStoredLanguage = (): string => {
  const stored = localStorage.getItem(LANGUAGE_KEY)
  return stored || 'en-US'
}

export const setLanguage = (lang: string) => {
  localStorage.setItem(LANGUAGE_KEY, lang)
  i18n.changeLanguage(lang)
}

i18n
  .use(initReactI18next)
  .init({
    resources: {
      'en-US': { translation: enUS },
      'zh-CN': { translation: zhCN },
    },
    lng: getStoredLanguage(),
    fallbackLng: 'en-US',
    interpolation: {
      escapeValue: false,
    },
  })

export default i18n
