import { type Ref } from 'vue'
import type { TodoItem } from '~/utils/agents/todoParser'
import type { ActiveTool } from '~/types/agents'

interface ToolExecution {
  toolName: string
  filePath?: string
  command?: string
  pattern?: string
  detail?: string
  timestamp?: Date
}

interface ToolManagementParams {
  sessionTodos: Ref<Map<string, TodoItem[]>>
  sessionToolExecution: Ref<Map<string, ToolExecution | null>>
  activeTools: Ref<Map<string, ActiveTool[]>>
}

export function useToolManagement(params: ToolManagementParams) {
  const { sessionTodos, sessionToolExecution, activeTools } = params

  // Update session data methods
  const updateSessionTodos = (sessionId: string, todos: TodoItem[]) => {
    sessionTodos.value.set(sessionId, todos)
  }

  const updateSessionToolExecution = (sessionId: string, toolExecution: ToolExecution | null) => {
    sessionToolExecution.value.set(sessionId, toolExecution)
  }

  const clearSessionToolExecution = (sessionId: string) => {
    sessionToolExecution.value.delete(sessionId)
  }

  // Tool overlay management
  const addActiveTool = (sessionId: string, toolUse: any) => {
    const tools = activeTools.value.get(sessionId) || []
    const activeTool: ActiveTool = {
      id: toolUse.id,
      name: toolUse.name,
      input: toolUse.input,
      status: 'running',
      startTime: Date.now(),
      sessionId
    }
    tools.push(activeTool)
    activeTools.value.set(sessionId, tools)
  }

  const completeActiveTool = (sessionId: string, toolUseId: string, isError: boolean = false) => {
    const tools = activeTools.value.get(sessionId) || []
    const tool = tools.find(t => t.id === toolUseId)
    if (tool) {
      tool.status = isError ? 'error' : 'completed'
      tool.endTime = Date.now()
      activeTools.value.set(sessionId, [...tools])
    }
  }

  const removeActiveTool = (sessionId: string, toolId: string) => {
    const tools = activeTools.value.get(sessionId) || []
    const filtered = tools.filter(t => t.id !== toolId)
    activeTools.value.set(sessionId, filtered)
  }

  return {
    updateSessionTodos,
    updateSessionToolExecution,
    clearSessionToolExecution,
    addActiveTool,
    completeActiveTool,
    removeActiveTool
  }
}
