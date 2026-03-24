"use client"

import { useRef, useState } from "react"
import { Mic, Send, Square, X } from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"

interface MessageInputProps {
  onSend: (message: string) => void
  onVoiceSend: (audioFile: File) => void
  disabled?: boolean
}

export function MessageInput({ onSend, onVoiceSend, disabled }: MessageInputProps) {
  const { t } = useTranslation()
  const [message, setMessage] = useState("")
  const [isRecording, setIsRecording] = useState(false)
  const [audioFile, setAudioFile] = useState<File | null>(null)
  const mediaRecorderRef = useRef<MediaRecorder | null>(null)
  const audioChunksRef = useRef<Blob[]>([])

  const handleSend = () => {
    if (audioFile && !disabled) {
      onVoiceSend(audioFile)
      setAudioFile(null)
      return
    }

    const nextMessage = message.trim()
    if (nextMessage && !disabled) {
      onSend(nextMessage)
      setMessage("")
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const startRecording = async () => {
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
        const audioBlob = new Blob(audioChunksRef.current, { type: "audio/webm" })
        const file = new File([audioBlob], "recording.webm", { type: "audio/webm" })
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
        <Textarea
          value={audioFile ? t("chat.audioReady", "Audio ready to send") : message}
          onChange={(e) => setMessage(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={t("chat.inputPlaceholder")}
          className="min-h-10 max-h-32 resize-none rounded-xl border border-border/60 bg-background/85 shadow-none focus-visible:ring-1"
          disabled={disabled || isRecording || Boolean(audioFile)}
          readOnly={Boolean(audioFile)}
        />

        {audioFile ? (
          <Button onClick={() => setAudioFile(null)} disabled={disabled} variant="ghost" size="icon" className="h-10 w-10 rounded-xl">
            <X className="h-4 w-4" />
          </Button>
        ) : (
          <Button
            onClick={isRecording ? stopRecording : startRecording}
            disabled={disabled}
            variant={isRecording ? "destructive" : "ghost"}
            size="icon"
            className="h-10 w-10 rounded-xl"
          >
            {isRecording ? <Square className="h-4 w-4" /> : <Mic className="h-4 w-4" />}
          </Button>
        )}

        <Button
          onClick={handleSend}
          disabled={disabled || (!message.trim() && !audioFile) || isRecording}
          className="h-10 w-10 rounded-xl"
          size="icon"
        >
          <Send className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
