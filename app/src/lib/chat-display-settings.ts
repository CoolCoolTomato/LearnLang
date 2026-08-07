const CHAT_DISPLAY_SETTINGS_KEY = "learnlang-chat-display-settings"

export interface ChatDisplaySettings {
  showVoiceChat: boolean
  showTextChat: boolean
}

const defaults: ChatDisplaySettings = {
  showVoiceChat: true,
  showTextChat: true,
}

export function getChatDisplaySettings(): ChatDisplaySettings {
  try {
    const stored = localStorage.getItem(CHAT_DISPLAY_SETTINGS_KEY)
    if (!stored) return defaults

    const parsed = JSON.parse(stored) as Partial<ChatDisplaySettings>
    const settings = {
      showVoiceChat: parsed.showVoiceChat !== false,
      showTextChat: parsed.showTextChat !== false,
    }
    if (!settings.showVoiceChat && !settings.showTextChat) {
      localStorage.setItem(CHAT_DISPLAY_SETTINGS_KEY, JSON.stringify(defaults))
      return defaults
    }
    return settings
  } catch {
    return defaults
  }
}

export function updateChatDisplaySettings(
  updates: Partial<ChatDisplaySettings>
): ChatDisplaySettings {
  const settings = { ...getChatDisplaySettings(), ...updates }
  if (!settings.showVoiceChat && !settings.showTextChat) return getChatDisplaySettings()

  localStorage.setItem(CHAT_DISPLAY_SETTINGS_KEY, JSON.stringify(settings))
  return settings
}
