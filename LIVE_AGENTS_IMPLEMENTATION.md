# Live Agents Implementation Plan

## Overview
Add real-time TodoWrite and tool execution tracking to the agents chat interface with **session-specific** tracking for multiple concurrent agent sessions.

## Problem Statement
Currently, when running agent sessions, users cannot see:
1. TodoWrite tasks being tracked in real-time
2. What tool is being executed (Bash, Read, Write, Edit, etc.)
3. Tool execution details (file paths, commands, etc.)

This makes it difficult to understand what the agent is doing, especially with multiple concurrent sessions.

## Solution
Implement two UI components that update in real-time:
1. **TodoWrite Box** (top-right) - Shows tasks with status
2. **Tool Execution Bar** (above input) - Shows current tool being executed

Both components are **session-specific**, so each open session shows its own state independently.

## Implementation Tasks

### 1. Frontend State Management
- [ ] Add reactive state variables for session-based tracking
- [ ] Add computed properties for active session data
- [ ] Create TypeScript interfaces for data structures

### 2. WebSocket Event Handlers
- [ ] Update `onAgentToolUse` to capture TodoWrite and tool execution
- [ ] Update `onAgentMessage` to clear tool execution
- [ ] Update `endSession` to clean up session data

### 3. UI Components
- [ ] Add TodoWrite box component to template
- [ ] Add Tool Execution bar component to template
- [ ] Add helper methods for formatting and icons

### 4. Styling
- [ ] Add CSS styles for TodoWrite box with animations
- [ ] Add CSS styles for Tool Execution bar with pulse animation

### 5. Testing
- [ ] Test multi-session TodoWrite tracking and switching
- [ ] Test tool execution bar with various tools

## Data Structures

### TodoItem Interface
```typescript
interface TodoItem {
  content: string              // Task description
  status: 'pending' | 'in_progress' | 'completed'
  activeForm?: string         // Present continuous form if applicable
}
```

### ToolExecution Interface
```typescript
interface ToolExecution {
  toolName: string            // 'Bash', 'Read', 'Write', 'Edit', 'Search'
  filePath?: string           // For file operations
  command?: string            // For Bash
  pattern?: string            // For Search
  timestamp: Date             // When tool use started
}
```

## UI Component Specifications

### TodoWrite Box (Top-Right)
- **Location**: Absolute positioned within `.chat-content`, top-right corner
- **Visibility**: Only when active session has todos
- **Features**:
  - Up to 10 todos, scrollable if more
  - Status icons (✅, 🔄, 📝)
  - Smooth fade-in/fade-out animations
  - Auto-hide when all todos completed

### Tool Execution Bar (Above Input)
- **Location**: Between `.messages-container` and `.input-area`
- **Visibility**: Only when tool is executing
- **Features**:
  - Tool-specific icons and formatting
  - File path truncation for long paths
  - Pulse animation while active
  - Auto-clear on message arrival

## Session Isolation

### Key Principles:
- Each session maintains its own todos and tool state
- Switching sessions updates UI to show that session's data
- Ending session cleans up its data
- Computed properties automatically re-render when active session changes

### Session Flow:
1. **New Session**: Empty todos and no active tool
2. **Switch Session**: Computed properties update automatically
3. **End Session**: Clean up session data from Maps

## File to Modify

`/Users/schlunsen/projects/claude-control-terminal/internal/server/frontend/app/pages/agents.vue`

## Testing Checklist

- [ ] Create session A → send prompt with TodoWrite → box appears
- [ ] Multiple todos display correctly with status icons
- [ ] Create session B → send prompt with TodoWrite → switch to A → correct todos show
- [ ] Switch to B → correct todos show
- [ ] Mark all todos complete → box auto-hides after 2s
- [ ] Run Bash command → tool bar appears with command
- [ ] Run Read file → tool bar shows file path
- [ ] Run Write file → tool bar shows file path
- [ ] Tool bar clears when message arrives
- [ ] Switch sessions while tool executing → tool bar updates
- [ ] End session → todos/tools cleared
- [ ] Long file paths truncated correctly in tool bar
- [ ] TodoWrite box scrollable with many todos
- [ ] Animations smooth and timing correct
- [ ] Multiple concurrent sessions each show correct state
- [ ] Mobile responsive (box and bar)

## Performance Considerations

- Using Maps for O(1) lookup by session ID
- Computed properties only update when activeSessionId changes
- No unnecessary re-renders
- Cleanup on session end prevents memory leaks
- Fade animations use CSS (GPU accelerated)

## Future Enhancements

- Click todo to copy content
- Click tool bar to see full details in modal
- Persist completed todos for session history
- Tool execution timing (show how long it took)
- Stack multiple simultaneous tools (if possible)