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
  let currentModel: string | null = null

  // Model size mapping
  const modelSizeMap: Record<string, string> = {
    'tiny': 'onnx-community/whisper-tiny.en',
    'base': 'onnx-community/whisper-base.en',
    'small': 'onnx-community/whisper-small.en'
  }

  /**
   * Fetch the selected Whisper model from settings
   */
  async function getSelectedModel(): Promise<string> {
    try {
      // Use authenticated fetch
      const { fetchWithAuth } = useAuthenticatedFetch()
      const response = await fetchWithAuth('/api/settings/whisper_model', {
        method: 'GET',
      })

      if (response.ok) {
        const setting = await response.json()
        const modelSize = setting.value || 'tiny'
        return modelSizeMap[modelSize] || modelSizeMap['tiny']
      }
    } catch (error) {
      console.error('Failed to fetch whisper model setting, using default:', error)
    }

    // Default to tiny model
    return modelSizeMap['tiny']
  }

  /**
   * Initialize the Whisper model
   * Model is selected based on user settings
   */
  async function initializeModel(): Promise<void> {
    try {
      isModelLoading.value = true
      error.value = null
      transcriptionProgress.value = 0 // Reset progress

      // Get the selected model from settings
      const selectedModel = await getSelectedModel()

      // If model is already loaded and it's the same model, skip reloading
      if (transcriber && currentModel === selectedModel) {
        console.log('Whisper model already loaded:', selectedModel)
        isModelLoading.value = false
        return
      }

      // If switching models, clear the old transcriber
      if (transcriber && currentModel !== selectedModel) {
        console.log('Switching from', currentModel, 'to', selectedModel)
        transcriber = null
        currentModel = null
      }

      console.log('Loading Whisper model from HuggingFace:', selectedModel)
      console.log('Model config:', {
        allowRemoteModels: env.allowRemoteModels,
        allowLocalModels: env.allowLocalModels,
        useBrowserCache: env.useBrowserCache
      })

      // Use onnx-community models which have better CORS support for web usage
      // These models are specifically optimized for browser environments

      // Track downloaded files for progress calculation
      const downloadedFiles = new Set<string>()
      const fileProgress = new Map<string, number>()

      transcriber = await pipeline(
        'automatic-speech-recognition',
        selectedModel,
        {
          quantized: true, // Use quantized version for faster loading
          progress_callback: (progress: any) => {
            console.log('Model loading progress:', progress)

            if (progress.status === 'progress' && progress.file) {
              // Store individual file progress
              fileProgress.set(progress.file, progress.progress || 0)

              // Calculate overall progress as average of all files
              const progressValues = Array.from(fileProgress.values())
              const avgProgress = progressValues.reduce((a, b) => a + b, 0) / Math.max(progressValues.length, 1)
              const percent = Math.min(Math.round(avgProgress * 100), 100)

              transcriptionProgress.value = percent
              console.log(`Downloading model: ${percent}% (file: ${progress.file})`)
            } else if (progress.status === 'download' && progress.file) {
              console.log(`Starting download: ${progress.file}`)
            } else if (progress.status === 'done' && progress.file) {
              downloadedFiles.add(progress.file)
              fileProgress.set(progress.file, 1.0)
              console.log(`Completed: ${progress.file}`)
            }
          }
        }
      )

      // Store the current model and mark as loaded
      currentModel = selectedModel
      transcriptionProgress.value = 100 // Ensure progress shows complete
      console.log('Whisper model loaded successfully:', selectedModel)
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

  /**
   * Get the currently loaded model name
   */
  function getCurrentModel(): string | null {
    return currentModel
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
    isSupported,
    getCurrentModel
  }
}
