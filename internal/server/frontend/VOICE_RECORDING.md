# Voice Recording & Transcription Feature

This document describes the voice recording and automatic transcription feature integrated into the Claude Control Terminal chat interface.

## Overview

The voice recording feature allows users to record audio messages and have them automatically transcribed using OpenAI's Whisper model running entirely in the browser via [Transformers.js](https://huggingface.co/docs/transformers.js).

## Features

- 🎤 **Voice Recording**: Record audio directly from your microphone
- 🤖 **Local AI Transcription**: Uses Whisper Tiny model (39M parameters) running in browser
- 💾 **Model Caching**: Models are downloaded once and cached in IndexedDB for instant reuse
- 📝 **Auto-insertion**: Transcribed text is automatically added to the chat input
- 🎨 **Visual Feedback**: Animated recording indicators and progress bars
- 🔒 **Privacy-focused**: All processing happens locally in your browser

## Usage

### Basic Flow

1. Click the **microphone button** (🎤) next to the send button
2. Allow microphone access when prompted (first time only)
3. Click **Start Recording** in the modal
4. Speak your message
5. Click **Stop & Transcribe** when done
6. Wait for transcription (first time will download the model ~75MB)
7. Transcribed text appears in the chat input
8. Edit if needed and send

### First Time Setup

On first use, the Whisper model will be downloaded (~75MB):
- The model is downloaded from Hugging Face CDN
- Progress is shown with a progress bar
- Model is cached in browser's IndexedDB
- Subsequent uses load instantly from cache
- Cache persists across browser sessions

### Keyboard Shortcuts

- **Enter**: Send message (when not recording)
- **Shift+Enter**: New line in text input
- **ESC**: Cancel recording (when modal is open)

## Technical Details

### Architecture

```
Voice Input → MediaRecorder → Audio Blob
                                    ↓
                            Whisper Model (Transformers.js)
                                    ↓
                              Text Transcription
                                    ↓
                              Chat Input Field
```

### Components

#### 1. `useVoiceRecording` Composable
Location: `app/composables/useVoiceRecording.ts`

Handles audio recording using Web Audio API:
- Uses `MediaRecorder` API
- Supports pause/resume
- Configures audio for optimal Whisper performance (16kHz sample rate)
- Provides duration tracking
- Error handling for microphone access

#### 2. `useWhisperTranscription` Composable
Location: `app/composables/useWhisperTranscription.ts`

Manages AI transcription:
- Loads Whisper Tiny English model (`Xenova/whisper-tiny.en`)
- Handles audio format conversion
- Resamples audio to 16kHz (Whisper requirement)
- Reports loading progress
- Caches model in browser

#### 3. `ChatArea` Component
Location: `app/components/agents/ChatArea.vue`

UI integration:
- Recording modal with visual feedback
- Microphone button in input area
- Animated recording indicators
- Progress tracking
- Error display

### Model Information

**Default Model**: `Xenova/whisper-tiny.en`
- Size: ~75MB
- Parameters: 39M
- Language: English only
- Speed: Fast (real-time on modern browsers)
- Accuracy: Good for clear speech

**Alternative Models** (can be configured):
- `Xenova/whisper-base` - 74M params, better accuracy, slightly slower
- `Xenova/whisper-small` - 244M params, best accuracy, slower
- `Xenova/whisper-medium` - 769M params, excellent accuracy, much slower

### Browser Requirements

**Required APIs**:
- `MediaDevices.getUserMedia()` - Microphone access
- `MediaRecorder` - Audio recording
- `AudioContext` - Audio processing
- `IndexedDB` - Model caching
- `Web Workers` - Background transcription (via Transformers.js)

**Supported Browsers**:
- Chrome/Edge 87+
- Firefox 94+
- Safari 14.1+
- Opera 73+

**Not Supported**:
- Internet Explorer
- Older mobile browsers

### Performance

**First Time Use**:
- Model download: ~10-30 seconds (depending on connection)
- Model initialization: ~2-5 seconds
- Total: ~15-35 seconds

**Subsequent Uses**:
- Model load from cache: <1 second
- Transcription: ~1-3 seconds per 10 seconds of audio

**Resource Usage**:
- Disk space (IndexedDB): ~75MB
- Memory during transcription: ~200-400MB
- CPU: Moderate (uses Web Workers)

## Configuration

### Changing the Whisper Model

Edit `app/composables/useWhisperTranscription.ts`:

```typescript
// Line ~48: Change model identifier
transcriber = await pipeline(
  'automatic-speech-recognition',
  'Xenova/whisper-base',  // Change this
  { /* ... */ }
)
```

**Available models**:
- `Xenova/whisper-tiny.en` (default) - 39M, English only
- `Xenova/whisper-base` - 74M, multilingual
- `Xenova/whisper-small` - 244M, multilingual
- `Xenova/whisper-medium` - 769M, multilingual

### Audio Recording Settings

Edit `app/composables/useVoiceRecording.ts`:

```typescript
// Line ~35: Modify audio constraints
stream = await navigator.mediaDevices.getUserMedia({
  audio: {
    echoCancellation: true,  // Remove echo
    noiseSuppression: true,  // Remove background noise
    sampleRate: 16000        // Whisper optimal rate
  }
})
```

### Cache Configuration

Edit `app/composables/useWhisperTranscription.ts`:

```typescript
// Lines 4-9: Transformers.js settings
env.allowRemoteModels = true   // Allow downloading models
env.allowLocalModels = true    // Allow cached models
env.useBrowserCache = true     // Enable IndexedDB cache
env.useFS = false              // Disable filesystem (browser)
```

## Troubleshooting

### Model Won't Download
- Check internet connection
- Check browser console for CORS errors
- Try clearing IndexedDB: Browser DevTools → Application → IndexedDB → Clear
- Verify CDN is accessible: https://cdn-lfs.huggingface.co/

### Microphone Not Working
- Check browser permissions (lock icon in address bar)
- Verify microphone is connected and working in system settings
- Try different browser
- Check for browser extensions blocking microphone access

### Transcription Inaccurate
- Speak clearly and at normal pace
- Reduce background noise
- Use a better microphone
- Consider upgrading to `whisper-base` or `whisper-small` model
- Ensure good internet connection (for first-time model download)

### Performance Issues
- Close other tabs/applications
- Use `whisper-tiny.en` for fastest speed
- Check available memory (needs ~400MB free)
- Try disabling browser extensions
- Update to latest browser version

### Cache Not Working
- Check available disk space (needs ~75MB)
- Verify IndexedDB is enabled in browser settings
- Clear and re-download: DevTools → Application → IndexedDB → Delete database
- Try incognito mode to test (will re-download model)

## Privacy & Security

### Data Handling
- ✅ All audio processing happens **locally in your browser**
- ✅ Audio is **never sent to external servers**
- ✅ Transcription happens **entirely client-side**
- ✅ No cloud API calls for transcription
- ⚠️ Initial model download requires internet connection

### Permissions
- **Microphone**: Required for audio recording
- **IndexedDB**: Required for model caching (improves performance)

### Data Storage
- **Audio recordings**: Temporary, deleted after transcription
- **AI models**: Cached in IndexedDB, can be cleared anytime
- **Transcribed text**: Only stored if you send the message

## Development

### Testing Locally

```bash
# Start dev server
cd internal/server/frontend
npm run dev

# Navigate to agents page
open http://localhost:3001/agents

# Create a session and test recording
```

### Building for Production

```bash
# Build frontend
npm run generate

# Models will be downloaded on first use in production
# Caching will work normally
```

### Debugging

Enable verbose logging:

```typescript
// In useWhisperTranscription.ts
console.log('Transcription result:', result)
console.log('Model loaded:', transcriber)

// In useVoiceRecording.ts
console.log('Recording state:', voiceRecording.isRecording.value)
console.log('Audio blob size:', audioBlob.value?.size)
```

## Future Enhancements

Potential improvements:
- [ ] Support for multilingual transcription
- [ ] Real-time streaming transcription
- [ ] Noise reduction preprocessing
- [ ] Voice activity detection (auto-stop)
- [ ] Audio waveform visualization
- [ ] Model selection in UI
- [ ] Offline mode indicator
- [ ] Transcription confidence scores
- [ ] Alternative transcription engines (browser APIs)

## Credits

- **Transformers.js**: [@xenova/transformers](https://github.com/xenova/transformers.js)
- **Whisper Model**: OpenAI
- **Whisper Web Demo**: [Hugging Face Space](https://huggingface.co/spaces/Xenova/whisper-web)

## License

This feature is part of Claude Control Terminal and follows the project's MIT license.
