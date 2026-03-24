import * as React from "react"
import { useNavigate } from "react-router-dom"
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
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import {
  Settings2,
  Bot,
  Brain,
  Mic,
  Volume2,
  Save,
  ArrowLeft,
  SettingsIcon,
} from "lucide-react"

import type { Settings, ProviderType, Language, UpdateSettingsRequest } from "@/types/settings"
import type { Model } from "@/types/model"

import { getCustomProviderModels } from "@/api/model-provider"
import { getSettings, updateSettings } from "@/api/settings"
import { setLanguage } from "@/i18n"
import { ThemeToggle } from "@/components/theme/theme-toggle"
import { ThemeColorSelect } from "@/components/theme/theme-color-select"
import { getErrorMessage } from "@/lib/error"

import { SettingsSection } from "./components/settings-section"
import { Field } from "./components/field"
import { ProviderModelSection } from "./components/provider-model-section"

export default function Page() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()

  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [activeTab, setActiveTab] = React.useState("general")
  const [language, setLanguageState] = React.useState<Language>(
    (i18n.resolvedLanguage || "zh-CN") as Language
  )


  const [models, setModels] = React.useState<Model[]>([])
  const [embeddingModels, setEmbeddingModels] = React.useState<Model[]>([])
  const [sttModels, setSttModels] = React.useState<Model[]>([])
  const [ttsModels, setTtsModels] = React.useState<Model[]>([])

  const [loadingModels, setLoadingModels] = React.useState(false)
  const [loadingEmbeddingModels, setLoadingEmbeddingModels] = React.useState(false)
  const [loadingSttModels, setLoadingSttModels] = React.useState(false)
  const [loadingTtsModels, setLoadingTtsModels] = React.useState(false)

  const [settingsFormData, setSettingsFormData] = React.useState({
    api_base_url: "",
    api_key: "",
    model: "",

    embedding_api_base_url: "",
    embedding_api_key: "",
    embedding_model: "",

    stt_api_base_url: "",
    stt_api_key: "",
    stt_model: "",

    tts_api_base_url: "",
    tts_api_key: "",
    tts_model: "",
    tts_voice: "",

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

  const languageList: Language[] = ["zh-CN", "en-US"]

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
      toast.error(getErrorMessage(err, "Failed to load models"))
      setModelState([])
    } finally {
      setLoadingState(false)
    }
  }

  const load = React.useCallback(async () => {
    try {
      setLoading(true)

      const [settingsData] = await Promise.all([
        getSettings(),
      ])

      const s = settingsData as Settings

      setSettingsFormData({
        api_base_url: s.api_base_url || "",
        api_key: s.api_key || "",
        model: s.model || "",

        embedding_api_base_url: s.embedding_api_base_url || "",
        embedding_api_key: s.embedding_api_key || "",
        embedding_model: s.embedding_model || "",

        stt_api_base_url: s.stt_api_base_url || "",
        stt_api_key: s.stt_api_key || "",
        stt_model: s.stt_model || "",

        tts_api_base_url: s.tts_api_base_url || "",
        tts_api_key: s.tts_api_key || "",
        tts_model: s.tts_model || "",
        tts_voice: s.tts_voice || "",

        native_language: s.native_language || "",
        target_language: s.target_language || "",
        timezone: s.timezone || "",
      })
      setLanguageState((i18n.resolvedLanguage || "zh-CN") as Language)

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
      const payload: UpdateSettingsRequest = {
        api_base_url: settingsFormData.api_base_url || undefined,
        api_key: settingsFormData.api_key || undefined,
        model: settingsFormData.model || undefined,
        embedding_api_base_url: settingsFormData.embedding_api_base_url || undefined,
        embedding_api_key: settingsFormData.embedding_api_key || undefined,
        embedding_model: settingsFormData.embedding_model || undefined,
        stt_api_base_url: settingsFormData.stt_api_base_url || undefined,
        stt_api_key: settingsFormData.stt_api_key || undefined,
        stt_model: settingsFormData.stt_model || undefined,
        tts_api_base_url: settingsFormData.tts_api_base_url || undefined,
        tts_api_key: settingsFormData.tts_api_key || undefined,
        tts_model: settingsFormData.tts_model || undefined,
        tts_voice: settingsFormData.tts_voice || undefined,
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
        <div className="text-sm text-muted-foreground">{t("common.loading")}</div>
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
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/chat")}
            className="h-8 w-15 text-muted-foreground"
            aria-label={t("common.back")}
            title={t("common.back")}
          >
            <ArrowLeft className="h-5 w-5" />
            <span className="ml-0.5">{t("common.back")}</span>
          </Button>
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full mt-2">
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

            <div className="bg-muted/50 w-full h-13 flex items-center justify-center rounded-3xl p-2 mt-8">
              <TabsList
                className="
                  justify-start!
                  flex w-full items-center gap-2
                  group-data-horizontal/tabs:h-10
                  overflow-x-auto overflow-y-hidden
                  whitespace-nowrap
                  bg-transparent p-0.5
                  rounded-none
                "
              >
                <TabsTrigger
                  value="general"
                  className="shrink-0 rounded-xl px-4"
                >
                  <Settings2 className="mr-2 h-4 w-4" />
                  {t("settings.generalTitle", "General")}
                </TabsTrigger>

                <TabsTrigger
                  value="chat"
                  className="shrink-0 rounded-xl px-4"
                >
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

                <TabsTrigger
                  value="stt"
                  className="shrink-0 rounded-xl px-4"
                >
                  <Mic className="mr-2 h-4 w-4" />
                  STT
                </TabsTrigger>

                <TabsTrigger
                  value="tts"
                  className="shrink-0 rounded-xl px-4"
                >
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

                <div className="grid gap-4 md:grid-cols-2">
                  <Field label={t("settings.nativeLanguage")}>
                    <Select
                      value={settingsFormData.native_language}
                      onValueChange={(value: ProviderType) => patchForm({ native_language: value })}
                    >
                      <SelectTrigger className="h-12">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="0">{t("settings.noProvider")}</SelectItem>
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
                      onValueChange={(value: ProviderType) => patchForm({ target_language: value })}
                    >
                      <SelectTrigger className="h-12">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="0">{t("settings.noProvider")}</SelectItem>
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
                    patchForm({ embedding_api_base_url: value, embedding_model: "" })
                    setEmbeddingModels([])
                  }}
                  onApiKeyChange={(value) => {
                    patchForm({ embedding_api_key: value, embedding_model: "" })
                    setEmbeddingModels([])
                  }}
                  onModelChange={(value) => patchForm({ embedding_model: value })}
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
                    (settingsFormData.tts_api_base_url &&
                        settingsFormData.tts_api_key) && (
                      <Field label={t("settings.ttsVoice")}>
                        <Input
                          value={settingsFormData.tts_voice}
                          onChange={(e) => patchForm({ tts_voice: e.target.value })}
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
