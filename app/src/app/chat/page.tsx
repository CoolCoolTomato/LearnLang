"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { toast } from "sonner"
import { useTranslation } from "react-i18next"
import { MessageList } from "./components/message-list"
import { MessageInput } from "./components/message-input"
import { ChatHeader } from "./components/chat-header"
import { getChatHistory, sendChatMessage, sendVoiceMessage } from "@/api/chat"
import type { ChatMessage } from "@/types/chat"
import { API_CONFIG, TOKEN_KEY } from "@/api/config"
import { getErrorMessage } from "@/lib/error"

export default function ChatPage() {
  const { t } = useTranslation()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [sending, setSending] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [wsConnected, setWsConnected] = useState(false)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const isManualCloseRef = useRef(false)

  const connectWebSocket = useCallback(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) return

    const wsUrl =
      API_CONFIG.BASE_URL.replace("http", "ws") +
      API_CONFIG.API_PREFIX +
      `/ws/chat?token=${token}`

    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      reconnectAttemptsRef.current = 0
      setWsConnected(true)
      console.log("WebSocket connected")
    }

    ws.onmessage = (event) => {
      try {
        const aiMessage: ChatMessage = JSON.parse(event.data)
        setMessages((prev) => [...prev, aiMessage])
      } catch (err) {
        console.error("Failed to parse WebSocket message:", err)
      }
    }

    ws.onerror = (error) => {
      console.error("WebSocket error:", error)
      setWsConnected(false)
    }

    ws.onclose = (event) => {
      console.log("WebSocket closed:", event.code, event.reason)
      setWsConnected(false)

      if (!isManualCloseRef.current && event.code !== 1000) {
        if (reconnectAttemptsRef.current >= 3) {
          toast.error(t("chat.wsConnectionFailed"))
          reconnectAttemptsRef.current = 0
          setWsConnected(false)
          return
        }

        reconnectAttemptsRef.current += 1
        reconnectTimeoutRef.current = setTimeout(() => {
          connectWebSocket()
        }, 2000)
      }
    }

    wsRef.current = ws
  }, [t])

  const loadInitialMessages = useCallback(async () => {
    try {
      setLoading(true)
      const response = await getChatHistory()
      setMessages(response.data || [])
      setHasMore((response.data?.length || 0) === 20)
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("chat.loadMessagesFailed")))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadInitialMessages()
    connectWebSocket()

    return () => {
      isManualCloseRef.current = true
      setWsConnected(false)

      if (wsRef.current) {
        wsRef.current.close()
      }

      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
    }
  }, [connectWebSocket, loadInitialMessages])

  const loadMoreMessages = async () => {
    if (loadingMore || !hasMore || messages.length === 0) return

    try {
      setLoadingMore(true)
      const oldestMessageId = messages[0].id
      const response = await getChatHistory({ before_id: oldestMessageId })
      setMessages((prev) => [...(response.data || []), ...prev])
      setHasMore((response.data?.length || 0) === 20)
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("chat.loadMessagesFailed")))
    } finally {
      setLoadingMore(false)
    }
  }

  const handleSend = async (message: string) => {
    try {
      setSending(true)
      const userMessage = await sendChatMessage({ message })
      setMessages((prev) => [...prev, userMessage])
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("chat.sendFailed")))
    } finally {
      setSending(false)
    }
  }

  const handleVoiceSend = async (audioFile: File) => {
    try {
      setSending(true)
      const userMessage = await sendVoiceMessage(audioFile)
      setMessages((prev) => [...prev, userMessage])
    } catch (err: unknown) {
      toast.error(getErrorMessage(err, t("chat.sendFailed")))
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-background">
      <ChatHeader
        title="LearnLang"
        subtitle={t("chat.headerSubtitle", "Your AI language partner")}
        connected={wsConnected}
      />

      <div className="min-h-0 flex-1">
        <MessageList
          messages={messages}
          loading={loading}
          loadingMore={loadingMore}
          hasMore={hasMore}
          onLoadMore={loadMoreMessages}
        />
      </div>

      <div className="border-t border-border/40 bg-gradient-to-t from-background via-background/95 to-background/70">
        <MessageInput
          onSend={handleSend}
          onVoiceSend={handleVoiceSend}
          disabled={sending}
        />
      </div>
    </div>
  )
}
