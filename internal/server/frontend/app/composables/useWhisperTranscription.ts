import { ref } from 'vue'
import { pipeline, env } from '@huggingface/transformers'

// Configure transformers.js environment
env.allowLocalModels = false // Don't use local models
env.allowRemoteModels = true // Allow downloading from HuggingFace
env.useBrowserCache = true // Cache in IndexedDB

console.log('Transformers.js environment configured for browser')

export interface TranscriptionResult {
  text: string
  chunks?: Array<{
    text: string
    timestamp: [number, number]
  }>
}

export function useWhisperTranscription() {
  const isTranscribing = ref(false)
  const isModelLoading = ref(false)
  const transcriptionProgress = ref(0)
  const error = ref<string | null>(null)
  const transcription = ref<string>('')

  let transcriber: any = null

  /**
   * Initialize the Whisper model
   * Using Whisper Tiny for faster inference on client-side
   */
  async function initializeModel(): Promise<void> {
    if (transcriber) return // Already initialized

    try {
      isModelLoading.value = true
      error.value = null

      console.log('Loading Whisper model from HuggingFace...')
      console.log('Model config:', {
        allowRemoteModels: env.allowRemoteModels,
        allowLocalModels: env.allowLocalModels,
        useBrowserCache: env.useBrowserCache
      })

      // Use onnx-community models which have better CORS support for web usage
      // These models are specifically optimized for browser environments
      transcriber = await pipeline(
        'automatic-speech-recognition',
        'onnx-community/whisper-tiny.en',
        {
          quantized: true, // Use quantized version for faster loading
          progress_callback: (progress: any) => {
            console.log('Model loading progress:', progress)
            if (progress.status === 'progress') {
              const percent = Math.round((progress.progress || 0) * 100)
              transcriptionProgress.value = percent
              console.log(`Downloading model: ${percent}%`)
            } else if (progress.status === 'download') {
              console.log(`Downloading: ${progress.name || progress.file}`)
            }
          }
        }
      )

      console.log('Whisper model loaded successfully')
    } catch (err) {
      console.error('Failed to initialize Whisper model:', err)
      // Log the full error with stack trace
      if (err instanceof Error) {
        console.error('Error details:', {
          message: err.message,
          stack: err.stack,
          name: err.name
        })
      }
      error.value = err instanceof Error ? err.message : 'Failed to load transcription model'

      // More user-friendly error messages
      if (err instanceof Error && err.message.includes('<!DOCTYPE')) {
        error.value = 'Failed to download model from HuggingFace. Please check your internet connection and try again.'
      }

      throw err
    } finally {
      isModelLoading.value = false
    }
  }

  /**
   * Convert audio blob to the format expected by Whisper
   */
  async function audioToFloat32Array(audioBlob: Blob): Promise<Float32Array> {
    // Create audio context
    const audioContext = new AudioContext({ sampleRate: 16000 })

    try {
      // Convert blob to array buffer
      const arrayBuffer = await audioBlob.arrayBuffer()

      // Decode audio data
      const audioBuffer = await audioContext.decodeAudioData(arrayBuffer)

      // Get channel data (mono)
      let audioData = audioBuffer.getChannelData(0)

      // Resample to 16kHz if needed (Whisper expects 16kHz)
      if (audioBuffer.sampleRate !== 16000) {
        audioData = resampleAudio(audioData, audioBuffer.sampleRate, 16000)
      }

      return audioData
    } finally {
      // Close audio context to free resources
      await audioContext.close()
    }
  }

  /**
   * Simple audio resampling
   */
  function resampleAudio(
    audioData: Float32Array,
    origSampleRate: number,
    targetSampleRate: number
  ): Float32Array {
    if (origSampleRate === targetSampleRate) {
      return audioData
    }

    const ratio = origSampleRate / targetSampleRate
    const newLength = Math.round(audioData.length / ratio)
    const result = new Float32Array(newLength)

    for (let i = 0; i < newLength; i++) {
      const srcIndex = Math.round(i * ratio)
      result[i] = audioData[srcIndex]
    }

    return result
  }

  /**
   * Transcribe audio blob using Whisper
   */
  async function transcribe(audioBlob: Blob): Promise<string> {
    try {
      isTranscribing.value = true
      error.value = null
      transcription.value = ''

      // Initialize model if not already loaded
      if (!transcriber) {
        await initializeModel()
      }

      console.log('Converting audio for transcription...')
      const audioData = await audioToFloat32Array(audioBlob)

      console.log('Transcribing audio...')
      const result = await transcriber(audioData, {
        chunk_length_s: 30,
        stride_length_s: 5,
        return_timestamps: false
      })

      transcription.value = result.text.trim()
      console.log('Transcription complete:', transcription.value)

      return transcription.value
    } catch (err) {
      console.error('Transcription failed:', err)
      error.value = err instanceof Error ? err.message : 'Transcription failed'
      throw err
    } finally {
      isTranscribing.value = false
    }
  }

  /**
   * Check if browser supports required APIs
   */
  function isSupported(): boolean {
    return !!(
      navigator.mediaDevices &&
      navigator.mediaDevices.getUserMedia &&
      window.AudioContext
    )
  }

  return {
    // State
    isTranscribing,
    isModelLoading,
    transcriptionProgress,
    error,
    transcription,

    // Methods
    initializeModel,
    transcribe,
    isSupported
  }
}
