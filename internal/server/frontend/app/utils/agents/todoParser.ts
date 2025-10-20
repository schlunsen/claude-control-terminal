/**
 * TodoWrite parsing utilities for agent conversations
 */

export interface TodoItem {
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  activeForm?: string
}

/**
 * Parse TodoWrite content from message
 * @param content - Raw message content
 * @returns Array of TodoItem objects or null if no todos found
 */
export const parseTodoWrite = (content: string): TodoItem[] | null => {
  if (!content || typeof content !== 'string') return null

  try {
    console.log('Parsing TodoWrite from content:', content)

    // Pattern 1: Look for numbered lists (1. task, 2. task, etc.)
    const numberedListMatch = content.match(/(?:\d+\.\s+)([^\n]+)/g)
    if (numberedListMatch) {
      console.log('Found numbered list matches:', numberedListMatch)
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
        console.log('Successfully parsed todos from numbered list:', todos)
        return todos
      }
    }

    // Pattern 2: Look for checkmark-style lists (- task, * task, etc.)
    const bulletListMatch = content.match(/[-*]\s+([^\n]+)/g)
    if (bulletListMatch) {
      console.log('Found bullet list matches:', bulletListMatch)
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
        console.log('Successfully parsed todos from bullet list:', todos)
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
        console.log('Successfully parsed todos from todo marker:', todos)
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
          console.log('Successfully parsed todos from task pattern:', todos)
          return todos
        }
      }
    }

    console.log('No TodoWrite data found in content')
    return null
  } catch (e) {
    console.warn('Failed to parse TodoWrite content:', e)
    return null
  }
}

/**
 * Format todos for tool use
 * @param todos - Array of TodoItem objects
 * @returns Formatted todos array for API
 */
export const formatTodosForTool = (todos: TodoItem[]): any[] => {
  return todos.map(todo => ({
    content: todo.content,
    status: todo.status,
    activeForm: todo.activeForm
  }))
}
