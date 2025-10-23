import { ref, onUnmounted } from 'vue'

export interface VoiceRecordingState {
  isRecording: boolean
  isPaused: boolean
  duration: number
  audioBlob: Blob | null
  error: string | null
}

export function useVoiceRecording() {
  const isRecording = ref(false)
  const isPaused = ref(false)
  const duration = ref(0)
  const audioBlob = ref<Blob | null>(null)
  const error = ref<string | null>(null)

  let mediaRecorder: MediaRecorder | null = null
  let audioChunks: Blob[] = []
  let stream: MediaStream | null = null
  let durationInterval: ReturnType<typeof setInterval> | null = null

  /**
   * Start recording audio from the microphone
   */
  async function startRecording(): Promise<void> {
    try {
      error.value = null
      audioChunks = []

      // Request microphone access
      stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          sampleRate: 16000 // Whisper prefers 16kHz
        }
      })

      // Create MediaRecorder with webm format (widely supported)
      const options = { mimeType: 'audio/webm' }
      mediaRecorder = new MediaRecorder(stream, options)

      // Collect audio data
      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunks.push(event.data)
        }
      }

      // Handle recording stop
      mediaRecorder.onstop = () => {
        const blob = new Blob(audioChunks, { type: 'audio/webm' })
        audioBlob.value = blob
        stopDurationTimer()
      }

      // Handle errors
      mediaRecorder.onerror = (event: Event) => {
        console.error('MediaRecorder error:', event)
        error.value = 'Recording error occurred'
        stopRecording()
      }

      // Start recording
      mediaRecorder.start()
      isRecording.value = true
      startDurationTimer()

    } catch (err) {
      console.error('Failed to start recording:', err)
      error.value = err instanceof Error ? err.message : 'Failed to access microphone'
      cleanup()
    }
  }

  /**
   * Stop recording
   */
  function stopRecording(): void {
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
      mediaRecorder.stop()
    }

    isRecording.value = false
    isPaused.value = false

    // Stop all tracks
    if (stream) {
      stream.getTracks().forEach(track => track.stop())
      stream = null
    }
  }

  /**
   * Pause recording
   */
  function pauseRecording(): void {
    if (mediaRecorder && mediaRecorder.state === 'recording') {
      mediaRecorder.pause()
      isPaused.value = true
      stopDurationTimer()
    }
  }

  /**
   * Resume recording
   */
  function resumeRecording(): void {
    if (mediaRecorder && mediaRecorder.state === 'paused') {
      mediaRecorder.resume()
      isPaused.value = false
      startDurationTimer()
    }
  }

  /**
   * Cancel recording (discard audio)
   */
  function cancelRecording(): void {
    stopRecording()
    audioChunks = []
    audioBlob.value = null
    duration.value = 0
  }

  /**
   * Reset state
   */
  function reset(): void {
    audioBlob.value = null
    duration.value = 0
    error.value = null
  }

  /**
   * Start duration timer
   */
  function startDurationTimer(): void {
    durationInterval = setInterval(() => {
      duration.value += 1
    }, 1000)
  }

  /**
   * Stop duration timer
   */
  function stopDurationTimer(): void {
    if (durationInterval) {
      clearInterval(durationInterval)
      durationInterval = null
    }
  }

  /**
   * Cleanup resources
   */
  function cleanup(): void {
    stopRecording()
    stopDurationTimer()
    audioChunks = []
    mediaRecorder = null
  }

  /**
   * Format duration as MM:SS
   */
  function formatDuration(seconds: number): string {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  // Cleanup on unmount
  onUnmounted(() => {
    cleanup()
  })

  return {
    // State
    isRecording,
    isPaused,
    duration,
    audioBlob,
    error,

    // Methods
    startRecording,
    stopRecording,
    pauseRecording,
    resumeRecording,
    cancelRecording,
    reset,
    formatDuration
  }
}
