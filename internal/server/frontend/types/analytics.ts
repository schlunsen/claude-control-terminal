// Analytics data types matching the Go backend

export interface ActivityItem {
  id: number
  type: 'shell' | 'claude' | 'prompt' | 'notification'
  timestamp: Date | string
  session_name?: string
  conversation_id: string
  git_branch?: string
  working_directory?: string

  // Shell command fields
  command?: string
  exit_code?: number
  stdout?: string
  stderr?: string
  duration_ms?: number

  // Claude command fields
  tool_name?: string
  parameters?: string
  result?: string
  success?: boolean
  error_message?: string

  // User prompt fields
  message?: string
  message_length?: number
  submitted_at?: string

  // Notification fields
  notification_type?: string
  command_details?: string
  notified_at?: string
}

export interface Stats {
  totalConversations: number
  totalTokens: number
  activeConversations: number
  avgTokens: number
  timestamp: string
  resetActive?: boolean
  resetTimestamp?: string
  resetReason?: string
}

export interface NotificationStats {
  total_notifications: number
  permission_requests: number
  idle_alerts: number
  most_requested_tool: string
  most_requested_tool_count: number
}

export interface WebSocketMessage {
  event: string
  data: any
  timestamp: string
}

export interface Process {
  PID: number
  Command: string
  WorkingDir: string
  Status: string
}

export interface Shell {
  shell_id?: string
  pid: number
  command: string
  working_dir?: string
  status: string
}
