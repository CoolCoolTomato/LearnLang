"use client"

import { useEffect, useRef, useState } from "react"
import { Languages, LoaderCircle, Mic, Send, Square, X } from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { translateText } from "@/api/translation"
import { getErrorMessage } from "@/lib/error"

const TRANSLATE_DEBOUNCE_MS = 250

interface MessageInputProps {
  onSend: (message: string) => void
  onVoiceSend: (audioFile: File) => void
  disabled?: boolean
}

export function MessageInput({
  onSend,
  onVoiceSend,
  disabled,
}: MessageInputProps) {
  const { t } = useTranslation()
  const [message, setMessage] = useState("")
  const [isRecording, setIsRecording] = useState(false)
  const [isTranslating, setIsTranslating] = useState(false)
  const [translationLocked, setTranslationLocked] = useState(false)
  const [audioFile, setAudioFile] = useState<File | null>(null)
  const mediaRecorderRef = useRef<MediaRecorder | null>(null)
  const audioChunksRef = useRef<Blob[]>([])
  const translationRequestRef = useRef(false)
  const translationDebounceRef = useRef<ReturnType<typeof setTimeout> | null>(
    null
  )
  const mountedRef = useRef(true)

  useEffect(() => {
    mountedRef.current = true
    return () => {
      mountedRef.current = false
      if (translationDebounceRef.current) {
        clearTimeout(translationDebounceRef.current)
      }
    }
  }, [])

  const handleSend = () => {
    if (isTranslating || translationRequestRef.current) return

    if (audioFile && !disabled) {
      onVoiceSend(audioFile)
      setAudioFile(null)
      return
    }

    const nextMessage = message.trim()
    if (nextMessage && !disabled) {
      onSend(nextMessage)
      setMessage("")
      setTranslationLocked(false)
    }
  }

  const handleMessageChange = (value: string) => {
    setMessage(value)
    if (translationLocked) {
      setTranslationLocked(false)
    }
  }

  const handleTranslate = () => {
    const sourceText = message.trim()
    if (
      !sourceText ||
      disabled ||
      isRecording ||
      audioFile ||
      isTranslating ||
      translationLocked ||
      translationRequestRef.current
    ) {
      return
    }

    translationRequestRef.current = true
    setIsTranslating(true)
    translationDebounceRef.current = setTimeout(async () => {
      try {
        const response = await translateText({ text: sourceText })
        if (!mountedRef.current) return
        setMessage(response.translation)
        setTranslationLocked(true)
      } catch (error: unknown) {
        if (mountedRef.current) {
          toast.error(getErrorMessage(error, t("chat.translateFailed")))
        }
      } finally {
        translationRequestRef.current = false
        translationDebounceRef.current = null
        if (mountedRef.current) {
          setIsTranslating(false)
        }
      }
    }, TRANSLATE_DEBOUNCE_MS)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const startRecording = async () => {
    if (isTranslating || translationRequestRef.current) return

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const mediaRecorder = new MediaRecorder(stream)
      mediaRecorderRef.current = mediaRecorder
      audioChunksRef.current = []

      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunksRef.current.push(event.data)
        }
      }

      mediaRecorder.onstop = () => {
        const audioBlob = new Blob(audioChunksRef.current, {
          type: "audio/webm",
        })
        const file = new File([audioBlob], "recording.webm", {
          type: "audio/webm",
        })
        setAudioFile(file)
        stream.getTracks().forEach((track) => track.stop())
      }

      mediaRecorder.start()
      setIsRecording(true)
    } catch {
      toast.error(t("chat.microphoneError", "Cannot access microphone"))
    }
  }

  const stopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop()
      setIsRecording(false)
    }
  }

  return (
    <div className="p-3 md:p-4">
      <div className="mx-auto flex w-full max-w-4xl items-end gap-2 rounded-2xl p-2">
        <Button
          type="button"
          onClick={handleTranslate}
          disabled={
            disabled ||
            !message.trim() ||
            isRecording ||
            Boolean(audioFile) ||
            isTranslating ||
            translationLocked
          }
          variant="ghost"
          size="icon"
          className="h-10 w-10 shrink-0 rounded-xl"
          aria-label={t("chat.translate")}
          title={t("chat.translate")}
        >
          {isTranslating ? (
            <LoaderCircle className="h-4 w-4 animate-spin" />
          ) : (
            <Languages className="h-4 w-4" />
          )}
        </Button>

        <Textarea
          value={
            audioFile ? t("chat.audioReady", "Audio ready to send") : message
          }
          onChange={(e) => handleMessageChange(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={t("chat.inputPlaceholder")}
          className="max-h-32 min-h-10 resize-none rounded-xl border border-border/60 bg-background/85 shadow-none focus-visible:ring-1"
          disabled={
            disabled || isRecording || Boolean(audioFile) || isTranslating
          }
          readOnly={Boolean(audioFile)}
        />

        {audioFile ? (
          <Button
            onClick={() => setAudioFile(null)}
            disabled={disabled}
            variant="ghost"
            size="icon"
            className="h-10 w-10 rounded-xl"
          >
            <X className="h-4 w-4" />
          </Button>
        ) : (
          <Button
            onClick={isRecording ? stopRecording : startRecording}
            disabled={disabled || isTranslating}
            variant={isRecording ? "destructive" : "ghost"}
            size="icon"
            className="h-10 w-10 rounded-xl"
          >
            {isRecording ? (
              <Square className="h-4 w-4" />
            ) : (
              <Mic className="h-4 w-4" />
            )}
          </Button>
        )}

        <Button
          onClick={handleSend}
          disabled={
            disabled ||
            (!message.trim() && !audioFile) ||
            isRecording ||
            isTranslating
          }
          className="h-10 w-10 rounded-xl"
          size="icon"
        >
          <Send className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
