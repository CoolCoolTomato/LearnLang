"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { Loader2, Pause, Play } from "lucide-react"
import { Button } from "@/components/ui/button"
import { getVoiceFileAudio } from "@/api/voice-file"
import type { VoiceFileInMessage } from "@/types/chat"

interface VoicePlayerProps {
  voiceFile: VoiceFileInMessage
  role?: "user" | "assistant"
}

export function VoicePlayer({
  voiceFile,
  role = "assistant",
}: VoicePlayerProps) {
  const [isPlaying, setIsPlaying] = useState(false)
  const [audioUrl, setAudioUrl] = useState<string>()
  const [loading, setLoading] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(voiceFile.duration ?? 0)

  const audioRef = useRef<HTMLAudioElement>(null)
  const rafRef = useRef<number | null>(null)

  const isUser = role === "user"

  useEffect(() => {
    return () => {
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current)
      }
      if (audioUrl) {
        URL.revokeObjectURL(audioUrl)
      }
    }
  }, [audioUrl])

  const formatDuration = (seconds: number) => {
    const safeSeconds = Math.max(0, Math.floor(seconds))
    const minute = Math.floor(safeSeconds / 60)
    const second = safeSeconds % 60
    return `${minute}:${String(second).padStart(2, "0")}`
  }

  const stopProgressLoop = useCallback(() => {
    if (rafRef.current !== null) {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
    }
  }, [])

  const startProgressLoop = useCallback(() => {
    stopProgressLoop()

    const update = () => {
      const audio = audioRef.current
      if (!audio) return

      setCurrentTime(audio.currentTime)

      if (!audio.paused && !audio.ended) {
        rafRef.current = requestAnimationFrame(update)
      } else {
        rafRef.current = null
      }
    }

    rafRef.current = requestAnimationFrame(update)
  }, [stopProgressLoop])

  const togglePlay = async () => {
    if (loading) return

    const audio = audioRef.current
    if (!audio) return

    if (isPlaying) {
      audio.pause()
      return
    }

    if (audioUrl) {
      await audio.play()
      return
    }

    try {
      setLoading(true)
      const blob = await getVoiceFileAudio(voiceFile.id)
      const url = URL.createObjectURL(blob)
      setAudioUrl(url)

      requestAnimationFrame(async () => {
        try {
          await audioRef.current?.play()
        } catch (err) {
          console.error("Failed to play voice file:", err)
        }
      })
    } catch (err) {
      console.error("Failed to load voice file:", err)
    } finally {
      setLoading(false)
    }
  }

  const progress =
    duration > 0 ? Math.min(currentTime / duration, 1) : 0

  return (
    <div className="mb-2 w-full">
      <div
        className={`flex w-full items-center gap-2.5 rounded-xl px-2.5 py-1.5 ${
          isUser
            ? "bg-primary-foreground/12"
            : "bg-black/5 dark:bg-white/5"
        }`}
      >
        <Button
          variant="ghost"
          size="icon"
          className={`h-7 w-7 rounded-full ${
            isUser
              ? "text-primary-foreground hover:bg-primary-foreground/15"
              : "text-foreground/75 hover:bg-background/60"
          }`}
          onClick={togglePlay}
          disabled={loading}
          aria-label={isPlaying ? "Pause voice message" : "Play voice message"}
        >
          {loading ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : isPlaying ? (
            <Pause className="h-4 w-4" />
          ) : (
            <Play className="h-4 w-4" />
          )}
        </Button>

        <div className="flex min-w-0 flex-1 items-center gap-2">
          <div className="flex h-4 items-end gap-0.5">
            <span
              className={`w-1 rounded-full transition-all ${
                isUser ? "bg-primary-foreground/85" : "bg-foreground/55"
              } ${isPlaying ? "h-3.5 motion-safe:animate-pulse" : "h-1.5"}`}
              style={{ animationDelay: "0ms" }}
            />
            <span
              className={`w-1 rounded-full transition-all ${
                isUser ? "bg-primary-foreground/85" : "bg-foreground/55"
              } ${isPlaying ? "h-4 motion-safe:animate-pulse" : "h-2.5"}`}
              style={{ animationDelay: "120ms" }}
            />
            <span
              className={`w-1 rounded-full transition-all ${
                isUser ? "bg-primary-foreground/85" : "bg-foreground/55"
              } ${isPlaying ? "h-2.5 motion-safe:animate-pulse" : "h-1"}`}
              style={{ animationDelay: "240ms" }}
            />
          </div>

          <div
            className={`h-0.5 flex-1 overflow-hidden rounded-full ${
              isUser ? "bg-primary-foreground/35" : "bg-foreground/20"
            }`}
          >
            <div
              className={`h-full origin-left rounded-full ${
                isUser ? "bg-primary-foreground" : "bg-foreground/70"
              }`}
              style={{
                transform: `scaleX(${progress})`,
                willChange: "transform",
              }}
            />
          </div>
        </div>

        <span
          className={`shrink-0 text-[10px] tabular-nums ${
            isUser ? "text-primary-foreground/75" : "text-muted-foreground"
          }`}
        >
          {formatDuration(duration > 0 ? duration - currentTime : voiceFile.duration ?? 0)}
        </span>
      </div>

      <audio
        ref={audioRef}
        src={audioUrl}
        preload="metadata"
        onLoadedMetadata={(event) => {
          const nextDuration = event.currentTarget.duration
          if (Number.isFinite(nextDuration) && nextDuration > 0) {
            setDuration(nextDuration)
          }
        }}
        onPlay={() => {
          setIsPlaying(true)
          startProgressLoop()
        }}
        onPause={() => {
          setIsPlaying(false)
          stopProgressLoop()
          const audio = audioRef.current
          if (audio) {
            setCurrentTime(audio.currentTime)
          }
        }}
        onEnded={() => {
          setIsPlaying(false)
          stopProgressLoop()
          setCurrentTime(0)
        }}
      />
    </div>
  )
}