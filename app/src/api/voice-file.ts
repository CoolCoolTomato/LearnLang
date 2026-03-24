import { API_CONFIG, TOKEN_KEY } from './config'

export const getVoiceFileAudio = async (id: number): Promise<Blob> => {
  const url = `${API_CONFIG.BASE_URL}${API_CONFIG.API_PREFIX}/voice-files/${id}/content`
  const response = await fetch(url, {
    headers: {
      'Authorization': `Bearer ${localStorage.getItem(TOKEN_KEY)}`
    }
  })
  if (!response.ok) throw new Error('Failed to load voice file')
  return await response.blob()
}
