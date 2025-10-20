/**
 * Tool execution parsing utilities for agent conversations
 */

export interface ToolExecution {
  toolName: string
  filePath?: string
  command?: string
  pattern?: string
  detail?: string
  timestamp: Date
}

/**
 * Parse tool use information from message content
 * @param content - Raw message content
 * @returns ToolExecution object or null if no tool use found
 */
export const parseToolUse = (content: string): ToolExecution | null => {
  if (!content || typeof content !== 'string') return null

  try {
    // Look for tool use patterns
    const patterns = [
      /Using (\w+)/g,
      /(\w+)\s*\(/g, // Function calls
      /Running (\w+)/g,
      /Executing (\w+)/g
    ]

    for (const pattern of patterns) {
      const matches = [...content.matchAll(pattern)]
      if (matches.length > 0) {
        const toolName = matches[0][1]

        // Extract additional details based on tool type
        let filePath, command, patternStr

        if (toolName === 'Read' || toolName === 'Write' || toolName === 'Edit') {
          const fileMatch = content.match(/(?:file|path):\s*([^\s\n]+)/i)
          if (fileMatch) filePath = fileMatch[1]
        } else if (toolName === 'Bash') {
          const commandMatch = content.match(/(?:command|cmd):\s*([^\n]+)/i)
          if (commandMatch) command = commandMatch[1].trim()
        } else if (toolName === 'Search' || toolName === 'Grep') {
          const patternMatch = content.match(/(?:pattern|search):\s*([^\n]+)/i)
          if (patternMatch) patternStr = patternMatch[1].trim()
        }

        return {
          toolName,
          filePath,
          command,
          pattern: patternStr,
          timestamp: new Date()
        }
      }
    }

    return null
  } catch (e) {
    console.warn('Failed to parse tool use:', e)
    return null
  }
}

/**
 * Get icon for a tool name
 * @param toolName - Name of the tool
 * @returns Emoji icon for the tool
 */
export const getToolIcon = (toolName: string): string => {
  const iconMap: Record<string, string> = {
    'Bash': '⚡',
    'Read': '📖',
    'Write': '✏️',
    'Edit': '🔧',
    'Search': '🔍',
    'Grep': '🔍',
    'TodoWrite': '📋'
  }

  return iconMap[toolName] || '🛠️'
}
