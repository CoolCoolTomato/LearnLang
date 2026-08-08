import * as React from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ModelCombobox } from "@/components/ui/model-combobox"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import {
  Settings2,
  Bot,
  Brain,
  Mic,
  Volume2,
  Save,
  SettingsIcon,
} from "lucide-react"

import type {
  Settings,
  ProviderType,
  LLMType,
  TTSType,
  Language,
  UpdateSettingsRequest,
} from "@/types/settings"
import type { Model } from "@/types/model"

import { getCustomProviderModels } from "@/api/model-provider"
import { getSettings, updateSettings } from "@/api/settings"
import { setLanguage } from "@/i18n"
import { ThemeToggle } from "@/components/theme/theme-toggle"
import { ThemeColorSelect } from "@/components/theme/theme-color-select"
import { Switch } from "@/components/ui/switch"
import { getErrorMessage } from "@/lib/error"
import {
  getChatDisplaySettings,
  updateChatDisplaySettings,
} from "@/lib/chat-display-settings"

import { SettingsSection } from "./components/settings-section"
import { Field } from "./components/field"
import { ProviderModelSection } from "./components/provider-model-section"

export default function Page() {
  const { t, i18n } = useTranslation()

  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [activeTab, setActiveTab] = React.useState("general")
  const [language, setLanguageState] = React.useState<Language>(
    (i18n.resolvedLanguage || "en-US") as Language
  )
  const [chatDisplaySettings, setChatDisplaySettings] = React.useState(
    getChatDisplaySettings
  )

  const [models, setModels] = React.useState<Model[]>([])
  const [embeddingModels, setEmbeddingModels] = React.useState<Model[]>([])
  const [sttModels, setSttModels] = React.useState<Model[]>([])
  const [ttsModels, setTtsModels] = React.useState<Model[]>([])

  const [loadingModels, setLoadingModels] = React.useState(false)
  const [loadingEmbeddingModels, setLoadingEmbeddingModels] =
    React.useState(false)
  const [loadingSttModels, setLoadingSttModels] = React.useState(false)
  const [loadingTtsModels, setLoadingTtsModels] = React.useState(false)

  const [settingsFormData, setSettingsFormData] = React.useState({
    api_base_url: "",
    api_key: "",
    model: "",
    llm_type: "openai" as LLMType,

    embedding_api_base_url: "",
    embedding_api_key: "",
    embedding_model: "",
    embedding_dimension: "",

    stt_api_base_url: "",
    stt_api_key: "",
    stt_model: "",

    tts_api_base_url: "",
    tts_api_key: "",
    tts_model: "",
    tts_voice: "",
    tts_type: "openai" as TTSType,

    native_language: "",
    target_language: "",
    timezone: "",
  })

  const commonTimezones = [
    { id: "Asia/Shanghai" },
    { id: "Asia/Singapore" },
    { id: "Asia/Tokyo" },
    { id: "Asia/Seoul" },
    { id: "Asia/Hong_Kong" },
    { id: "America/New_York" },
    { id: "America/Los_Angeles" },
    { id: "America/Chicago" },
    { id: "Europe/London" },
    { id: "Europe/Paris" },
    { id: "UTC" },
  ]

  const languageList: Language[] = ["en-US", "zh-CN"]

  const patchForm = (patch: Partial<typeof settingsFormData>) => {
    setSettingsFormData((prev) => ({ ...prev, ...patch }))
  }

  const loadCustomModels = async (
    apiBaseUrl: string,
    apiKey: string,
    setLoadingState: React.Dispatch<React.SetStateAction<boolean>>,
    setModelState: React.Dispatch<React.SetStateAction<Model[]>>
  ) => {
    if (!apiBaseUrl || !apiKey) {
      setModelState([])
      return
    }

    try {
      setLoadingState(true)
      const response = await getCustomProviderModels({
        api_base_url: apiBaseUrl,
        api_key: apiKey,
      })
      setModelState(response.data || [])
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("systemSettings.loadModelsFailed")))
      setModelState([])
    } finally {
      setLoadingState(false)
    }
  }

  const load = React.useCallback(async () => {
    try {
      setLoading(true)

      const [settingsData] = await Promise.all([getSettings()])

      const s = settingsData as Settings

      setSettingsFormData({
        api_base_url: s.api_base_url || "",
        api_key: s.api_key || "",
        model: s.model || "",
        llm_type: s.llm_type === "anthropic" ? "anthropic" : "openai",

        embedding_api_base_url: s.embedding_api_base_url || "",
        embedding_api_key: s.embedding_api_key || "",
        embedding_model: s.embedding_model || "",
        embedding_dimension: s.embedding_dimension
          ? String(s.embedding_dimension)
          : "",

        stt_api_base_url: s.stt_api_base_url || "",
        stt_api_key: s.stt_api_key || "",
        stt_model: s.stt_model || "",

        tts_api_base_url: s.tts_api_base_url || "",
        tts_api_key: s.tts_api_key || "",
        tts_model: s.tts_model || "",
        tts_voice: s.tts_voice || "",
        tts_type: s.tts_type === "fish-audio" ? "fish-audio" : "openai",

        native_language: s.native_language || "",
        target_language: s.target_language || "",
        timezone: s.timezone || "",
      })
      setLanguageState((i18n.resolvedLanguage || "en-US") as Language)

      setError(null)
    } catch (err: unknown) {
      const message = getErrorMessage(err, t("user.loadFailed"))
      setError(message)
      toast.error(message)
    } finally {
      setLoading(false)
    }
  }, [t, i18n])

  React.useEffect(() => {
    load()
  }, [load])

  const handleSaveSettings = async () => {
    try {
      setSaving(true)
      const embeddingDimension = Number(settingsFormData.embedding_dimension)
      const hasEmbeddingConfig = Boolean(
        settingsFormData.embedding_api_base_url ||
        settingsFormData.embedding_api_key ||
        settingsFormData.embedding_model
      )
      if (
        hasEmbeddingConfig &&
        (!Number.isInteger(embeddingDimension) || embeddingDimension <= 0)
      ) {
        toast.error(t("settings.embeddingDimensionRequired"))
        return
      }
      const payload: UpdateSettingsRequest = {
        api_base_url: settingsFormData.api_base_url || undefined,
        api_key: settingsFormData.api_key || undefined,
        model: settingsFormData.model || undefined,
        llm_type: settingsFormData.llm_type,
        embedding_api_base_url:
          settingsFormData.embedding_api_base_url || undefined,
        embedding_api_key: settingsFormData.embedding_api_key || undefined,
        embedding_model: settingsFormData.embedding_model || undefined,
        embedding_dimension:
          embeddingDimension > 0 ? embeddingDimension : undefined,
        stt_api_base_url: settingsFormData.stt_api_base_url || undefined,
        stt_api_key: settingsFormData.stt_api_key || undefined,
        stt_model: settingsFormData.stt_model || undefined,
        tts_api_base_url: settingsFormData.tts_api_base_url || undefined,
        tts_api_key: settingsFormData.tts_api_key || undefined,
        tts_model: settingsFormData.tts_model || undefined,
        tts_voice: settingsFormData.tts_voice || undefined,
        tts_type: settingsFormData.tts_type,
        native_language: settingsFormData.native_language || undefined,
        target_language: settingsFormData.target_language || undefined,
        timezone: settingsFormData.timezone || undefined,
      }
      await updateSettings(payload)
      toast.success(t("settings.updateSuccess"))
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("userSettings.updateFailed")))
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center">
        <div className="text-sm text-muted-foreground">
          {t("common.loading")}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-2xl border border-destructive/20 bg-destructive/10 p-4 text-sm text-destructive">
        {error}
      </div>
    )
  }

  return (
    <div className="relative min-h-full bg-background">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-6 md:px-6 md:py-8">
        <div className="rounded-3xl border border-border/60 bg-background/80 p-6 shadow-sm backdrop-blur">
          <Tabs
            value={activeTab}
            onValueChange={setActiveTab}
            className="w-full"
          >
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div className="flex items-center gap-3">
                <div className="inline-flex h-12 w-12 items-center justify-center rounded-xl border border-border/60 bg-muted/40">
                  <SettingsIcon className="h-6 w-6 text-muted-foreground" />
                </div>
                <div>
                  <h1 className="text-2xl font-semibold tracking-tight">
                    {t("settings.title", "Settings")}
                  </h1>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {t(
                      "settings.description",
                      "Manage your language preferences, theme, and AI model providers."
                    )}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Select
                  value={language}
                  onValueChange={(value: Language) => {
                    setLanguageState(value)
                    setLanguage(value)
                  }}
                >
                  <SelectTrigger className="h-10 w-30">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {languageList.map((item) => (
                      <SelectItem key={item} value={item}>
                        {item === "zh-CN"
                          ? t("settings.language_zhCN")
                          : t("settings.language_enUS")}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <ThemeColorSelect />
                <ThemeToggle />
              </div>
            </div>

            <div className="mt-8 flex h-13 w-full items-center justify-center rounded-3xl bg-muted/50 p-2">
              <TabsList className="flex w-full items-center justify-start! gap-2 overflow-x-auto overflow-y-hidden rounded-none bg-transparent p-0.5 whitespace-nowrap group-data-horizontal/tabs:h-10">
                <TabsTrigger
                  value="general"
                  className="shrink-0 rounded-xl px-4"
                >
                  <Settings2 className="mr-2 h-4 w-4" />
                  {t("settings.generalTitle", "General")}
                </TabsTrigger>

                <TabsTrigger value="chat" className="shrink-0 rounded-xl px-4">
                  <Bot className="mr-2 h-4 w-4" />
                  Chat
                </TabsTrigger>

                <TabsTrigger
                  value="embedding"
                  className="shrink-0 rounded-xl px-4"
                >
                  <Brain className="mr-2 h-4 w-4" />
                  Embedding
                </TabsTrigger>

                <TabsTrigger value="stt" className="shrink-0 rounded-xl px-4">
                  <Mic className="mr-2 h-4 w-4" />
                  STT
                </TabsTrigger>

                <TabsTrigger value="tts" className="shrink-0 rounded-xl px-4">
                  <Volume2 className="mr-2 h-4 w-4" />
                  TTS
                </TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value="general" className="mt-6">
              <SettingsSection
                title={t("settings.generalTitle", "General")}
                description={t(
                  "settings.generalDescription",
                  "Basic appearance and language learning preferences."
                )}
              >
                <Field label={t("settings.timezone")}>
                  <ModelCombobox
                    value={settingsFormData.timezone}
                    onValueChange={(value) => patchForm({ timezone: value })}
                    models={commonTimezones}
                    placeholder="Asia/Singapore"
                    className="h-10"
                  />
                </Field>

                <div className="grid gap-3 rounded-xl border border-border/60 p-4">
                  <div className="flex items-center justify-between gap-4">
                    <div className="min-w-0">
                      <div className="text-sm font-medium">
                        {t("settings.showVoiceChat")}
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {t("settings.showVoiceChatDescription")}
                      </p>
                    </div>
                    <Switch
                      checked={chatDisplaySettings.showVoiceChat}
                      disabled={!chatDisplaySettings.showTextChat}
                      onCheckedChange={(showVoiceChat) =>
                        setChatDisplaySettings(
                          updateChatDisplaySettings({ showVoiceChat })
                        )
                      }
                      aria-label={t("settings.showVoiceChat")}
                    />
                  </div>
                  <div className="flex items-center justify-between gap-4">
                    <div className="min-w-0">
                      <div className="text-sm font-medium">
                        {t("settings.showTextChat")}
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {t("settings.showTextChatDescription")}
                      </p>
                    </div>
                    <Switch
                      checked={chatDisplaySettings.showTextChat}
                      disabled={!chatDisplaySettings.showVoiceChat}
                      onCheckedChange={(showTextChat) =>
                        setChatDisplaySettings(
                          updateChatDisplaySettings({ showTextChat })
                        )
                      }
                      aria-label={t("settings.showTextChat")}
                    />
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {t("settings.chatDisplayHint")}
                  </p>
                </div>

                <div className="grid gap-4 md:grid-cols-2">
                  <Field label={t("settings.nativeLanguage")}>
                    <Select
                      value={settingsFormData.native_language}
                      onValueChange={(value: ProviderType) =>
                        patchForm({ native_language: value })
                      }
                    >
                      <SelectTrigger className="h-12">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="0">
                          {t("settings.noProvider")}
                        </SelectItem>
                        {languageList.map((language) => (
                          <SelectItem key={language} value={language}>
                            {language}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </Field>
                  <Field label={t("settings.targetLanguage")}>
                    <Select
                      value={settingsFormData.target_language}
                      onValueChange={(value: ProviderType) =>
                        patchForm({ target_language: value })
                      }
                    >
                      <SelectTrigger className="h-12">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="0">
                          {t("settings.noProvider")}
                        </SelectItem>
                        {languageList.map((language) => (
                          <SelectItem key={language} value={language}>
                            {language}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </Field>
                </div>
              </SettingsSection>
            </TabsContent>

            <TabsContent value="chat" className="mt-6">
              <SettingsSection
                title={t("settings.chatModelTitle", "Chat Model")}
                description={t(
                  "settings.chatModelDescription",
                  "Used for the main conversation and AI replies."
                )}
              >
                <Field
                  label={t("settings.llmType")}
                  description={t("settings.llmTypeDescription")}
                >
                  <Select
                    value={settingsFormData.llm_type}
                    onValueChange={(value: LLMType) =>
                      patchForm({ llm_type: value })
                    }
                  >
                    <SelectTrigger className="h-11 rounded-xl">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="openai">
                        {t("settings.llmTypeOpenAI")}
                      </SelectItem>
                      <SelectItem value="anthropic">
                        {t("settings.llmTypeAnthropic")}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <ProviderModelSection
                  apiBaseUrl={settingsFormData.api_base_url}
                  apiKey={settingsFormData.api_key}
                  model={settingsFormData.model}
                  models={models}
                  loadingModels={loadingModels}
                  apiBaseUrlLabel={t("settings.apiBaseUrl")}
                  apiKeyLabel={t("settings.apiKey")}
                  modelLabel={t("settings.model")}
                  onApiBaseUrlChange={(value) => {
                    patchForm({ api_base_url: value, model: "" })
                    setModels([])
                  }}
                  onApiKeyChange={(value) => {
                    patchForm({ api_key: value, model: "" })
                    setModels([])
                  }}
                  onModelChange={(value) => patchForm({ model: value })}
                  manualModelEntry={settingsFormData.llm_type === "anthropic"}
                  onLoadModels={() => {
                    if (
                      settingsFormData.api_base_url &&
                      settingsFormData.api_key
                    ) {
                      loadCustomModels(
                        settingsFormData.api_base_url,
                        settingsFormData.api_key,
                        setLoadingModels,
                        setModels
                      )
                    }
                  }}
                />
              </SettingsSection>
            </TabsContent>

            <TabsContent value="embedding" className="mt-6">
              <SettingsSection
                title={t("settings.embeddingSettings")}
                description={t(
                  "settings.embeddingDescription",
                  "Used for memory retrieval and semantic matching."
                )}
              >
                <ProviderModelSection
                  apiBaseUrl={settingsFormData.embedding_api_base_url}
                  apiKey={settingsFormData.embedding_api_key}
                  model={settingsFormData.embedding_model}
                  models={embeddingModels}
                  loadingModels={loadingEmbeddingModels}
                  apiBaseUrlLabel={t("settings.embeddingApiBaseUrl")}
                  apiKeyLabel={t("settings.embeddingApiKey")}
                  modelLabel={t("settings.embeddingModel")}
                  onApiBaseUrlChange={(value) => {
                    patchForm({
                      embedding_api_base_url: value,
                      embedding_model: "",
                    })
                    setEmbeddingModels([])
                  }}
                  onApiKeyChange={(value) => {
                    patchForm({ embedding_api_key: value, embedding_model: "" })
                    setEmbeddingModels([])
                  }}
                  onModelChange={(value) =>
                    patchForm({ embedding_model: value })
                  }
                  onLoadModels={() => {
                    if (
                      settingsFormData.embedding_api_base_url &&
                      settingsFormData.embedding_api_key
                    ) {
                      loadCustomModels(
                        settingsFormData.embedding_api_base_url,
                        settingsFormData.embedding_api_key,
                        setLoadingEmbeddingModels,
                        setEmbeddingModels
                      )
                    }
                  }}
                  extra={
                    <Field
                      label={t("settings.embeddingDimension")}
                      htmlFor="embedding-dimension"
                      description={t("settings.embeddingDimensionDescription")}
                    >
                      <Input
                        id="embedding-dimension"
                        type="number"
                        min={1}
                        step={1}
                        inputMode="numeric"
                        required
                        value={settingsFormData.embedding_dimension}
                        onChange={(event) =>
                          patchForm({ embedding_dimension: event.target.value })
                        }
                        className="h-11 rounded-xl"
                      />
                    </Field>
                  }
                />
              </SettingsSection>
            </TabsContent>

            <TabsContent value="stt" className="mt-6">
              <SettingsSection
                title={t("settings.sttSettings")}
                description={t(
                  "settings.sttDescription",
                  "Speech-to-text settings for voice transcription."
                )}
              >
                <ProviderModelSection
                  apiBaseUrl={settingsFormData.stt_api_base_url}
                  apiKey={settingsFormData.stt_api_key}
                  model={settingsFormData.stt_model}
                  models={sttModels}
                  loadingModels={loadingSttModels}
                  apiBaseUrlLabel={t("settings.sttApiBaseUrl")}
                  apiKeyLabel={t("settings.sttApiKey")}
                  modelLabel={t("settings.sttModel")}
                  onApiBaseUrlChange={(value) => {
                    patchForm({ stt_api_base_url: value, stt_model: "" })
                    setSttModels([])
                  }}
                  onApiKeyChange={(value) => {
                    patchForm({ stt_api_key: value, stt_model: "" })
                    setSttModels([])
                  }}
                  onModelChange={(value) => patchForm({ stt_model: value })}
                  onLoadModels={() => {
                    if (
                      settingsFormData.stt_api_base_url &&
                      settingsFormData.stt_api_key
                    ) {
                      loadCustomModels(
                        settingsFormData.stt_api_base_url,
                        settingsFormData.stt_api_key,
                        setLoadingSttModels,
                        setSttModels
                      )
                    }
                  }}
                />
              </SettingsSection>
            </TabsContent>

            <TabsContent value="tts" className="mt-6">
              <SettingsSection
                title={t("settings.ttsSettings")}
                description={t(
                  "settings.ttsDescription",
                  "Text-to-speech settings for AI voice playback."
                )}
              >
                <Field label={t("settings.ttsProviderType")}>
                  <Select
                    value={settingsFormData.tts_type}
                    onValueChange={(value: TTSType) =>
                      patchForm({ tts_type: value })
                    }
                  >
                    <SelectTrigger className="h-11 rounded-xl">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="openai">OpenAI Compatible</SelectItem>
                      <SelectItem value="fish-audio">Fish Audio</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <ProviderModelSection
                  apiBaseUrl={settingsFormData.tts_api_base_url}
                  apiKey={settingsFormData.tts_api_key}
                  model={settingsFormData.tts_model}
                  models={ttsModels}
                  loadingModels={loadingTtsModels}
                  apiBaseUrlLabel={t("settings.ttsApiBaseUrl")}
                  apiKeyLabel={t("settings.ttsApiKey")}
                  modelLabel={t("settings.ttsModel")}
                  onApiBaseUrlChange={(value) => {
                    patchForm({ tts_api_base_url: value, tts_model: "" })
                    setTtsModels([])
                  }}
                  onApiKeyChange={(value) => {
                    patchForm({ tts_api_key: value, tts_model: "" })
                    setTtsModels([])
                  }}
                  onModelChange={(value) => patchForm({ tts_model: value })}
                  onLoadModels={() => {
                    if (
                      settingsFormData.tts_api_base_url &&
                      settingsFormData.tts_api_key
                    ) {
                      loadCustomModels(
                        settingsFormData.tts_api_base_url,
                        settingsFormData.tts_api_key,
                        setLoadingTtsModels,
                        setTtsModels
                      )
                    }
                  }}
                  extra={
                    settingsFormData.tts_api_base_url &&
                    settingsFormData.tts_api_key && (
                      <Field label={t("settings.ttsVoice")}>
                        <Input
                          value={settingsFormData.tts_voice}
                          onChange={(e) =>
                            patchForm({ tts_voice: e.target.value })
                          }
                          placeholder={t("settings.ttsVoicePlaceholder")}
                          className="h-11 rounded-xl"
                        />
                      </Field>
                    )
                  }
                />
              </SettingsSection>
            </TabsContent>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="text-sm text-muted-foreground">
                {t(
                  "settings.bottomHint",
                  "Your changes will affect chat, memory retrieval, and voice features."
                )}
              </div>
              <Button
                onClick={handleSaveSettings}
                disabled={saving}
                className="h-11 rounded-xl px-5"
              >
                <Save className="mr-2 h-4 w-4" />
                {saving ? t("common.saving", "Saving...") : t("common.save")}
              </Button>
            </div>
          </Tabs>
        </div>
      </div>
    </div>
  )
}
