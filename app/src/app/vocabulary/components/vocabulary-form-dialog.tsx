import * as React from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { createVocabulary, updateVocabulary } from "@/api/vocabulary"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { getErrorMessage } from "@/lib/error"
import type { Vocabulary, VocabularySummary } from "@/types/vocabulary"

interface VocabularyFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  vocabulary: VocabularySummary | null
  onSaved: (vocabulary: Vocabulary) => Promise<void>
}

export function VocabularyFormDialog({
  open,
  onOpenChange,
  vocabulary,
  onSaved,
}: VocabularyFormDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = React.useState("")
  const [targetLanguage, setTargetLanguage] = React.useState("")
  const [nativeLanguage, setNativeLanguage] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState("")
  const editing = vocabulary !== null

  React.useEffect(() => {
    if (!open) return
    setName(vocabulary?.name ?? "")
    setTargetLanguage(vocabulary?.target_language ?? "")
    setNativeLanguage(vocabulary?.native_language ?? "")
    setError("")
  }, [open, vocabulary])

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!name.trim()) {
      setError(t("vocabulary.nameRequired"))
      return
    }
    try {
      setSaving(true)
      setError("")
      const payload = {
        name: name.trim(),
        target_language: targetLanguage.trim(),
        native_language: nativeLanguage.trim(),
      }
      const result = vocabulary
        ? await updateVocabulary(vocabulary.id, payload)
        : await createVocabulary(payload)
      await onSaved(result)
      toast.success(
        editing ? t("vocabulary.updateSuccess") : t("vocabulary.createSuccess")
      )
    } catch (saveError: unknown) {
      setError(getErrorMessage(saveError, t("vocabulary.saveFailed")))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => !saving && onOpenChange(next)}>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit} className="grid gap-4">
          <DialogHeader>
            <DialogTitle>
              {editing
                ? t("vocabulary.editTitle")
                : t("vocabulary.createTitle")}
            </DialogTitle>
            <DialogDescription>
              {editing
                ? t("vocabulary.editDescription")
                : t("vocabulary.createDescription")}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-2">
            <Label
              htmlFor={
                editing ? "edit-vocabulary-name" : "create-vocabulary-name"
              }
            >
              {t("vocabulary.name")}
            </Label>
            <Input
              id={editing ? "edit-vocabulary-name" : "create-vocabulary-name"}
              value={name}
              onChange={(event) => setName(event.target.value)}
              disabled={saving}
              autoFocus
            />
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <div className="grid gap-2">
              <Label
                htmlFor={
                  editing ? "edit-target-language" : "create-target-language"
                }
              >
                {t("settings.targetLanguage")}
              </Label>
              <Input
                id={editing ? "edit-target-language" : "create-target-language"}
                value={targetLanguage}
                onChange={(event) => setTargetLanguage(event.target.value)}
                placeholder="en-US"
                disabled={saving}
              />
            </div>
            <div className="grid gap-2">
              <Label
                htmlFor={
                  editing ? "edit-native-language" : "create-native-language"
                }
              >
                {t("settings.nativeLanguage")}
              </Label>
              <Input
                id={editing ? "edit-native-language" : "create-native-language"}
                value={nativeLanguage}
                onChange={(event) => setNativeLanguage(event.target.value)}
                placeholder="zh-CN"
                disabled={saving}
              />
            </div>
          </div>

          {error ? (
            <div className="border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          ) : null}

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" disabled={saving}>
                {t("common.cancel")}
              </Button>
            </DialogClose>
            <Button type="submit" disabled={saving}>
              {saving ? t("common.saving") : t("common.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
