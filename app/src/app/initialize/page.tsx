import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { ArrowRight, BookOpenText } from "lucide-react"
import { getSettings, updateSettings } from "@/api/settings"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

const timezones = ["Asia/Shanghai", "Asia/Singapore", "Asia/Tokyo", "Asia/Seoul", "America/New_York", "America/Los_Angeles", "Europe/London", "Europe/Paris", "UTC"]
const languages = [
  { value: "en", label: "English" },
  { value: "zh-CN", label: "Chinese" },
  { value: "ja", label: "Japanese" },
  { value: "ko", label: "Korean" },
  { value: "es", label: "Spanish" },
  { value: "fr", label: "French" },
]

export default function InitializationPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [timezone, setTimezone] = useState("")
  const [nativeLanguage, setNativeLanguage] = useState("")
  const [targetLanguage, setTargetLanguage] = useState("")
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    getSettings()
      .then((settings) => {
        setTimezone(settings.timezone || "")
        setNativeLanguage(settings.native_language || "")
        setTargetLanguage(settings.target_language || "")
      })
      .catch(() => setError(t("onboarding.loadFailed", "Unable to load your preferences.")))
      .finally(() => setLoading(false))
  }, [t])

  const save = async () => {
    if (!timezone || !nativeLanguage || !targetLanguage) {
      setError(t("onboarding.required", "Choose a timezone, native language, and learning language."))
      return
    }
    if (nativeLanguage === targetLanguage) {
      setError(t("onboarding.differentLanguages", "Choose different native and learning languages."))
      return
    }
    try {
      setSaving(true)
      setError("")
      await updateSettings({ timezone, native_language: nativeLanguage, target_language: targetLanguage })
      navigate("/chat", { replace: true })
    } catch {
      setError(t("onboarding.saveFailed", "Unable to save your preferences."))
    } finally {
      setSaving(false)
    }
  }

  return (
    <main className="flex min-h-dvh items-center justify-center bg-muted/30 p-6">
      {loading ? <LoadingSpinner /> : (
      <Card className="w-full max-w-xl shadow-sm">
        <CardHeader className="space-y-4">
          <div className="flex size-11 items-center justify-center rounded-xl bg-primary text-primary-foreground"><BookOpenText className="size-5" /></div>
          <div>
            <CardTitle className="text-2xl">{t("onboarding.title", "Set up your learning space")}</CardTitle>
            <CardDescription className="mt-2">{t("onboarding.description", "Tell us how you learn so LearnLang can tailor every conversation.")}</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="grid gap-5">
          <div className="grid gap-2"><Label>{t("onboarding.timezone", "Timezone")}</Label><Select value={timezone} onValueChange={setTimezone}><SelectTrigger className="w-full"><SelectValue placeholder={t("onboarding.chooseTimezone", "Choose your timezone")} /></SelectTrigger><SelectContent>{timezones.map((value) => <SelectItem key={value} value={value}>{value}</SelectItem>)}</SelectContent></Select></div>
          <div className="grid gap-2"><Label>{t("onboarding.nativeLanguage", "Native language")}</Label><Select value={nativeLanguage} onValueChange={setNativeLanguage}><SelectTrigger className="w-full"><SelectValue placeholder={t("onboarding.chooseLanguage", "Choose a language")} /></SelectTrigger><SelectContent>{languages.map((language) => <SelectItem key={language.value} value={language.value}>{language.label}</SelectItem>)}</SelectContent></Select></div>
          <div className="grid gap-2"><Label>{t("onboarding.targetLanguage", "Learning language")}</Label><Select value={targetLanguage} onValueChange={setTargetLanguage}><SelectTrigger className="w-full"><SelectValue placeholder={t("onboarding.chooseLanguage", "Choose a language")} /></SelectTrigger><SelectContent>{languages.map((language) => <SelectItem key={language.value} value={language.value}>{language.label}</SelectItem>)}</SelectContent></Select></div>
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <Button type="button" className="w-full" disabled={saving} onClick={save}>{saving ? t("onboarding.saving", "Saving...") : t("onboarding.continue", "Continue")}<ArrowRight /></Button>
        </CardContent>
      </Card>
      )}
    </main>
  )
}
