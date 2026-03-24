"use client"

import { useEffect, useRef, useState } from "react"
import { format, isToday, isYesterday } from "date-fns"
import { useTranslation } from "react-i18next"
import { User, Bot } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { cn } from "@/lib/utils"
import type { ChatMessage } from "@/types/chat"
import { VoicePlayer } from "./voice-player"
import { useAuth } from "@/contexts/auth-context"
import { resolveAvatarUrl } from "@/api/profile"

interface MessageListProps {
  messages: ChatMessage[]
  loading?: boolean
  loadingMore?: boolean
  hasMore?: boolean
  onLoadMore?: () => void
}

export function MessageList({ messages, loading, loadingMore, hasMore, onLoadMore }: MessageListProps) {
  const { t } = useTranslation()
  const { user } = useAuth()
  const bottomRef = useRef<HTMLDivElement>(null)
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const prevScrollHeightRef = useRef(0)
  const prevMessageCountRef = useRef(0)
  const [showTranslation, setShowTranslation] = useState<Record<number, boolean>>({})
  const timeoutRefs = useRef<Record<number, ReturnType<typeof setTimeout>>>({})

  const toggleTranslation = (messageId: number) => {
    setShowTranslation(prev => {
      const newState = !prev[messageId]

      if (timeoutRefs.current[messageId]) {
        clearTimeout(timeoutRefs.current[messageId])
      }

      if (newState) {
        timeoutRefs.current[messageId] = setTimeout(() => {
          setShowTranslation(prev => ({ ...prev, [messageId]: false }))
          delete timeoutRefs.current[messageId]
        }, 3000)
      }

      return { ...prev, [messageId]: newState }
    })
  }

  const formatDateLabel = (date: Date) => {
    if (isToday(date)) {
      return t("chat.today")
    }
    if (isYesterday(date)) {
      return t("chat.yesterday")
    }
    const locale = t("common.locale")
    if (locale === "en") {
      return format(date, "MMM d")
    }
    return format(date, "M月d日")
  }

  const shouldShowDateSeparator = (currentMessage: ChatMessage, prevMessage?: ChatMessage) => {
    if (!prevMessage) return true
    const currentDate = format(new Date(currentMessage.created_at), "yyyy-MM-dd")
    const prevDate = format(new Date(prevMessage.created_at), "yyyy-MM-dd")
    return currentDate !== prevDate
  }

  const getViewport = () => {
    return scrollAreaRef.current?.querySelector('[data-slot="scroll-area-viewport"]') as HTMLDivElement | null
  }

  useEffect(() => {
    const viewport = getViewport()
    if (!viewport) return

    const handleScrollEvent = () => {
      if (!hasMore || loadingMore) return
      if (viewport.scrollTop === 0) {
        prevScrollHeightRef.current = viewport.scrollHeight
        onLoadMore?.()
      }
    }

    viewport.addEventListener('scroll', handleScrollEvent)
    return () => viewport.removeEventListener('scroll', handleScrollEvent)
  }, [hasMore, loadingMore, onLoadMore])

  useEffect(() => {
    if (!loading && messages.length > 0) {
      bottomRef.current?.scrollIntoView({ behavior: "auto" })
    }
  }, [loading, messages.length])

  useEffect(() => {
    if (messages.length > prevMessageCountRef.current && !loadingMore) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" })
    }
    prevMessageCountRef.current = messages.length
  }, [messages.length, loadingMore])

  useEffect(() => {
    const viewport = getViewport()
    if (viewport && loadingMore === false && prevScrollHeightRef.current > 0) {
      const newScrollHeight = viewport.scrollHeight
      viewport.scrollTop = newScrollHeight - prevScrollHeightRef.current
      prevScrollHeightRef.current = 0
    }
  }, [loadingMore])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-muted-foreground">{t("chat.loading")}</div>
      </div>
    )
  }

  const sortedMessages = [...messages].sort((a, b) =>
    new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
  )

  return (
    <ScrollArea ref={scrollAreaRef} className="flex-1 h-full">
      <div className="p-4 min-h-full">
        {loadingMore && (
          <div className="text-center text-sm text-muted-foreground py-2">
            {t("chat.loading")}
          </div>
        )}
        <div className="space-y-4 md:px-1">
        {sortedMessages.map((message, index) => (
          <div key={message.id}>
            {shouldShowDateSeparator(message, sortedMessages[index - 1]) && (
              <div className="flex items-center justify-center my-6">
                <div className="text-xs text-muted-foreground bg-muted px-3 py-1 rounded-full">
                  {formatDateLabel(new Date(message.created_at))}
                </div>
              </div>
            )}
            <div
              className={cn(
                "flex gap-3",
                message.role === "user" ? "justify-end" : "justify-start"
              )}
            >
              {message.role === "assistant" && (
                <Avatar
                  className="w-10 h-10 cursor-pointer"
                  onClick={() => message.translation && toggleTranslation(message.id)}
                >
                  <AvatarFallback>
                    <Bot className="w-5 h-5" />
                  </AvatarFallback>
                </Avatar>
              )}
              <div
                className={cn(
                  "max-w-[70%] rounded-lg px-4 py-2",
                  message.role === "user"
                    ? "chat-user-bubble"
                    : "bg-muted"
                )}
              >
                {message.voice_file && (
                  <VoicePlayer voiceFile={message.voice_file} role={message.role === "user" ? "user" : "assistant"} />
                )}
                <div className="text-sm">
                  {showTranslation[message.id] && message.translation
                    ? message.translation
                    : message.text_content}
                </div>
                {message.created_at && (
                  <div className="text-xs opacity-70 mt-1">
                    {format(new Date(message.created_at), "HH:mm")}
                  </div>
                )}
              </div>
              {message.role === "user" && (
                <Avatar className="w-10 h-10">
                  <AvatarImage src={resolveAvatarUrl(user?.avatar_url)} alt={user?.username || "User"} />
                  <AvatarFallback>
                    <User className="w-5 h-5" />
                  </AvatarFallback>
                </Avatar>
              )}
            </div>
          </div>
        ))}
        <div ref={bottomRef} /></div>
      </div>
    </ScrollArea>
  )
}
