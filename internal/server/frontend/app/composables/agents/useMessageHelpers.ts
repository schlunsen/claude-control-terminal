import type { TodoItem } from '~/utils/agents/todoParser'

interface ToolExecution {
  toolName: string
  filePath?: string
  command?: string
  pattern?: string
  timestamp: Date
}

export function useMessageHelpers() {
  // Format relative time
  const formatRelativeTime = (timestamp: string | Date) => {
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

  // Helper methods for parsing TodoWrite and tool execution data
  const parseTodoWrite = (content: string): TodoItem[] | null => {
    if (!content || typeof content !== 'string') return null

    try {
      // Pattern 1: Look for numbered lists (1. task, 2. task, etc.)
      const numberedListMatch = content.match(/(?:\d+\.\s+)([^\n]+)/g)
      if (numberedListMatch) {
        const todos: TodoItem[] = []
        for (let i = 0; i < numberedListMatch.length; i++) {
          const taskContent = numberedListMatch[i].replace(/^\d+\.\s+/, '').trim()
          if (taskContent) {
            todos.push({
              content: taskContent,
              status: i === 0 ? 'in_progress' : 'pending'
            })
          }
        }
        if (todos.length > 0) {
          return todos
        }
      }

      // Pattern 2: Look for checkmark-style lists (- task, * task, etc.)
      const bulletListMatch = content.match(/[-*]\s+([^\n]+)/g)
      if (bulletListMatch) {
        const todos: TodoItem[] = []
        for (const match of bulletListMatch) {
          const taskContent = match.replace(/^[-*]\s+/, '').trim()
          if (taskContent) {
            todos.push({
              content: taskContent,
              status: 'pending'
            })
          }
        }
        if (todos.length > 0) {
          return todos
        }
      }

      // Pattern 3: Look for explicit todo markers ("Todo:", "Task:", etc.)
      const todoMarkerMatch = content.match(/(?:todo|task|items?):\s*([\s\S]*?)(?=\n\n|\n\w+:|$)/i)
      if (todoMarkerMatch) {
        const todoText = todoMarkerMatch[1].trim()
        const items = todoText.split(/\n\s*\n/).filter(item => item.trim())
        if (items.length > 0) {
          const todos = items.map(item => ({
            content: item.trim(),
            status: 'pending' as const
          }))
          return todos
        }
      }

      // Pattern 4: Look for task-related patterns (common task descriptions)
      const taskPatterns = [
        /(?:i'll create|let me create|creating|here are)\s+a\s+(?:todo|task|list):\s*([\s\S]*?)(?=\n\n|\n|$)/i,
        /(?:tasks?:\s*\n)((?:\d+\.\s+[^\n]+\n)+)/i,
        /(?:items?:\s*\n)((?:[-*]\s+[^\n]+\n)+)/i
      ]

      for (const pattern of taskPatterns) {
        const match = content.match(pattern)
        if (match) {
          const taskContent = match[1] || match[0]
          const lines = taskContent.split('\n').filter(line => line.trim())
          const todos = lines.map(line => ({
            content: line.trim().replace(/^\d+\.\s+/, '').replace(/^[-*]\s+/, ''),
            status: 'pending' as const
          })).filter(todo => todo.content)

          if (todos.length > 0) {
            return todos
          }
        }
      }

      return null
    } catch (e) {
      console.warn('Failed to parse TodoWrite content:', e)
      return null
    }
  }

  const parseToolUse = (content: string): ToolExecution | null => {
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

  // Helper to format todos for TodoWrite tool (includes activeForm only when present)
  const formatTodosForTool = (todos: TodoItem[]): any[] => {
    return todos.map(todo => ({
      content: todo.content,
      status: todo.status,
      ...(todo.activeForm && { activeForm: todo.activeForm })
    }))
  }

  const truncatePath = (path: string): string => {
    if (!path) return ''
    if (path.length <= 50) return path

    // Truncate from the middle, keeping the beginning and end
    const start = path.substring(0, 25)
    const end = path.substring(path.length - 20)
    return `${start}...${end}`
  }

  // Helper function to extract text content from nested content object
  const extractTextContent = (content: any): string => {
    if (!content) return ''

    // If content is already a string, return it
    if (typeof content === 'string') return content

    // If content is an object with nested structure
    if (typeof content === 'object') {
      // Handle assistant messages with text array
      if (content.type === 'assistant') {
        // Check if text array exists and is not null/empty
        if (Array.isArray(content.text) && content.text.length > 0) {
          return content.text.join('\n')
        }
        // Empty or null text array - no content to display
        return ''
      }

      // Handle user messages
      if (content.type === 'user' && content.content) {
        return String(content.content)
      }

      // Handle result messages (completion signal - no visible content)
      if (content.type === 'result') {
        return ''
      }

      // Handle system messages
      if (content.type === 'system') {
        return `SystemMessage: ${content.subtype || 'unknown'}`
      }

      // Fallback: stringify the object
      return JSON.stringify(content)
    }

    return String(content)
  }

  // Helper to check if content is complete signal
  const isCompleteSignal = (content: any): boolean => {
    return typeof content === 'object' && content.type === 'result'
  }

  // Helper to extract cost/usage data from result messages
  const extractCostData = (content: any) => {
    if (typeof content === 'object' && content.type === 'result') {
      return {
        costUSD: content.cost_usd || 0,
        numTurns: content.num_turns || 0,
        durationMs: content.duration_ms || 0,
        usage: content.usage || null
      }
    }
    return null
  }

  // Helper to extract tool name from tool_uses JSON
  const extractToolName = (toolUses: any): string | undefined => {
    if (!toolUses) return undefined

    try {
      // Parse if it's a JSON string
      const parsed = typeof toolUses === 'string' ? JSON.parse(toolUses) : toolUses

      // If it's an array, get the first tool
      if (Array.isArray(parsed) && parsed.length > 0) {
        return parsed[0].name || parsed[0].type
      }

      // If it's a single object
      if (parsed.name || parsed.type) {
        return parsed.name || parsed.type
      }
    } catch (e) {
      console.warn('Failed to parse tool_uses:', e)
    }

    return undefined
  }

  return {
    formatRelativeTime,
    parseTodoWrite,
    parseToolUse,
    formatTodosForTool,
    truncatePath,
    extractTextContent,
    isCompleteSignal,
    extractCostData,
    extractToolName
  }
}
