/**
 * Message and time formatting utilities for agent conversations
 */

/**
 * Format a timestamp to a readable time string
 * @param timestamp - Date timestamp
 * @returns Formatted time string (e.g., "2:30 PM")
 */
export const formatTime = (timestamp: string | number | Date): string => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  })
}

/**
 * Format a timestamp as relative time
 * @param timestamp - Date timestamp
 * @returns Relative time string (e.g., "5m ago", "2h ago")
 */
export const formatRelativeTime = (timestamp: string | number | Date): string => {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

/**
 * Format message content for display with markdown support
 * @param content - Raw message content
 * @returns HTML-formatted message string
 */
export const formatMessage = (content: any): string => {
  // If content is an object, extract text from it first
  if (typeof content === 'object' && content !== null) {
    // Try to extract meaningful text from the object
    if (content.text) {
      content = Array.isArray(content.text) ? content.text.join('\n') : String(content.text)
    } else if (content.content) {
      content = String(content.content)
    } else {
      // For other objects, create a concise representation
      const objType = content.type || 'unknown'
      const keys = Object.keys(content).filter(k => k !== 'type')
      if (keys.length === 0) {
        return `<em class="system-message">${objType}</em>`
      }
      // Show key properties in a readable format
      const props = keys.slice(0, 3).map(k => `${k}: ${String(content[k]).substring(0, 30)}`).join(', ')
      return `<em class="system-message">${objType} - ${props}</em>`
    }
  }

  // Ensure content is a string at this point
  content = String(content)

  // Skip system messages and JSON-like content
  if (content.includes('SystemMessage(') || (content.startsWith('{') && content.includes('"type"'))) {
    return '<em class="system-message">Processing...</em>'
  }

  // Clean up the content
  let cleanContent = content

  // If it's a string representation of an object, try to extract meaningful text
  if (cleanContent.includes('assistant:')) {
    const match = cleanContent.match(/assistant:\s*(.+?)(?:\n|$)/i)
    if (match) {
      cleanContent = match[1]
    }
  }

  // Convert markdown to HTML (basic)
  return cleanContent
    .replace(/```(.*?)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/\n/g, '<br>')
}

/**
 * Truncate a path or text to a maximum length
 * @param text - Text to truncate
 * @param maxLength - Maximum length
 * @returns Truncated text with ellipsis if needed
 */
export const truncatePath = (text: string | undefined | null, maxLength: number): string => {
  if (!text) return ''
  if (text.length <= maxLength) return text

  // For paths, try to keep the filename
  if (text.includes('/')) {
    const parts = text.split('/')
    const filename = parts[parts.length - 1]
    if (filename.length < maxLength - 3) {
      return `.../${filename}`
    }
  }

  return text.substring(0, maxLength - 3) + '...'
}
