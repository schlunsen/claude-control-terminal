# Notification Hook Implementation Plan

## Overview

Add Notification hook to capture permission requests and idle alerts, providing engagement metrics and trust evolution tracking in the analytics dashboard.

## Architecture

### Current State
- ✅ UserPromptSubmit hook captures user input
- ✅ PostToolUse hook captures tool usage
- ✅ Database tables for shell_commands, claude_commands, user_messages
- ✅ Analytics dashboard with activity history

### New Addition
- Notification hook captures:
  1. **Permission requests** - "Claude needs your permission to use Bash"
  2. **Idle alerts** - "Claude is waiting for your input" (after 60s idle)

## Implementation Steps

### 1. Create Notification Logger Hook Script

**File**: `hooks/notification-logger.sh`

**Features**:
- Hook Type: Notification (no matcher needed)
- Parse JSON input from stdin
- Extract fields:
  - `session_id` - conversation identifier
  - `message` - notification text
  - `cwd` - working directory (from environment)
  - `hook_event_name` - confirms "Notification"

**Notification Type Detection**:
```bash
if echo "$MESSAGE" | grep -qi "permission"; then
    NOTIFICATION_TYPE="permission_request"
    # Extract tool name from message (e.g., "use Bash" -> "Bash")
    TOOL_NAME=$(echo "$MESSAGE" | grep -oP '(?<=use )\w+' || echo "")
elif echo "$MESSAGE" | grep -qi "waiting.*input"; then
    NOTIFICATION_TYPE="idle_alert"
    TOOL_NAME=""
else
    NOTIFICATION_TYPE="other"
    TOOL_NAME=""
fi
```

**API Endpoint**:
- POST to `http://localhost:3333/api/notifications`

**Payload**:
```json
{
  "session_id": "uuid",
  "session_name": "Cartman",
  "notification_type": "permission_request|idle_alert|other",
  "message": "Claude needs your permission to use Bash",
  "tool_name": "Bash",
  "cwd": "/path/to/project",
  "branch": "main"
}
```

**Session Naming**:
- Use same hash-based approach as other hooks
- 10 South Park character names for consistency

**Execution**:
- Run curl/wget in background (non-blocking)
- Silent failures (don't interrupt Claude Code)

### 2. Database Schema Changes

**File**: `internal/database/schema.sql`

**New Table**: `notifications`
```sql
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL,
    session_name TEXT,
    notification_type TEXT NOT NULL, -- 'permission_request', 'idle_alert', 'other'
    message TEXT NOT NULL,
    tool_name TEXT, -- extracted from permission requests
    working_directory TEXT,
    git_branch TEXT,
    notified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_conversation
    ON notifications(conversation_id, notified_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_type
    ON notifications(notification_type, notified_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_tool
    ON notifications(tool_name, notified_at DESC) WHERE tool_name IS NOT NULL;
```

**Migration**: `internal/database/database.go`
```go
// Migration 5: Create notifications table
_, err := db.Exec(`CREATE TABLE IF NOT EXISTS notifications (...)`)
```

### 3. Database Models

**File**: `internal/database/models.go`

**New Model**:
```go
type Notification struct {
    ID               int64     `json:"id"`
    ConversationID   string    `json:"conversation_id"`
    SessionName      string    `json:"session_name,omitempty"`
    NotificationType string    `json:"notification_type"`
    Message          string    `json:"message"`
    ToolName         string    `json:"tool_name,omitempty"`
    WorkingDirectory string    `json:"working_directory,omitempty"`
    GitBranch        string    `json:"git_branch,omitempty"`
    NotifiedAt       time.Time `json:"notified_at"`
    CreatedAt        time.Time `json:"created_at"`
}
```

### 4. Repository Methods

**File**: `internal/database/repository.go`

**New Methods**:
```go
func (r *Repository) RecordNotification(notif *Notification) error
func (r *Repository) GetNotifications(query *CommandHistoryQuery) ([]*Notification, error)
func (r *Repository) GetNotificationStats() (*NotificationStats, error)
func (r *Repository) DeleteAllNotifications() error
```

**Stats Model**:
```go
type NotificationStats struct {
    TotalNotifications     int64 `json:"total_notifications"`
    PermissionRequests     int64 `json:"permission_requests"`
    IdleAlerts             int64 `json:"idle_alerts"`
    MostRequestedTool      string `json:"most_requested_tool"`
    MostRequestedToolCount int64 `json:"most_requested_tool_count"`
}
```

### 5. API Endpoints

**File**: `internal/server/server.go`

**New Endpoints**:
```go
// POST /api/notifications - Record notification
api.Post("/notifications", s.handleRecordNotification)

// GET /api/notifications - Get notification history
api.Get("/notifications", s.handleGetNotifications)

// GET /api/notifications/stats - Get notification statistics
api.Get("/notifications/stats", s.handleGetNotificationStats)

// DELETE /api/notifications - Clear all notifications
api.Delete("/notifications", s.handleClearNotifications)
```

**Update Unified History**:
```go
// GET /api/history/all - Include notifications
// Add notifications to the unified timeline
```

### 6. Hook Installer

**File**: `internal/components/hook.go`

**New Methods**:
```go
func (hi *HookInstaller) InstallNotificationLogger() error
func (hi *HookInstaller) UninstallNotificationLogger() error
func (hi *HookInstaller) CheckNotificationLoggerInstalled() (bool, error)
```

**Update InstallAllHooks**:
```go
func (hi *HookInstaller) InstallAllHooks() error {
    // Install user prompt logger
    // Install tool logger
    // Install notification logger (NEW)
}
```

**Settings Format**:
```json
{
  "hooks": {
    "UserPromptSubmit": [...],
    "PostToolUse": [...],
    "Notification": [
      {
        "hooks": [
          {"type": "command", "command": ".claude/hooks/notification-logger.sh"}
        ]
      }
    ]
  }
}
```

### 7. CLI Commands

**File**: `internal/cmd/root.go`

**New Flags**:
```go
var (
    installNotificationHook   bool
    uninstallNotificationHook bool
)

// Flags
rootCmd.Flags().BoolVar(&installNotificationHook, "install-notification-hook", false, "install notification logger hook (project-only)")
rootCmd.Flags().BoolVar(&uninstallNotificationHook, "uninstall-notification-hook", false, "uninstall notification logger hook")
```

**Update handleHookManagement**:
```go
// Handle notification hook install/uninstall
if installNotificationHook {
    if err := hookInstaller.InstallNotificationLogger(); err != nil {
        ShowError(fmt.Sprintf("Failed to install notification hook: %v", err))
        os.Exit(1)
    }
    return
}
```

**Update --install-all-hooks**:
- Now installs 3 hooks: user-prompt, tool, notification

### 8. Frontend Analytics Dashboard

**File**: `internal/server/static/index.html`

#### A. Add Notification Tab to Filter Tabs

```javascript
<div class="filter-tabs">
    <button class="filter-tab active" data-type="all" onclick="filterHistory('all')">All</button>
    <button class="filter-tab" data-type="shell" onclick="filterHistory('shell')">Shell</button>
    <button class="filter-tab" data-type="claude" onclick="filterHistory('claude')">Claude</button>
    <button class="filter-tab" data-type="prompt" onclick="filterHistory('prompt')">Prompts</button>
    <button class="filter-tab" data-type="notification" onclick="filterHistory('notification')">Notifications</button>
</div>
```

#### B. Update loadActivityHistory Function

```javascript
async function loadActivityHistory() {
    try {
        let url = `/api/history/all?limit=${HISTORY_PER_PAGE}&offset=0`;
        if (currentSessionFilter) {
            url += `&conversation_id=${encodeURIComponent(currentSessionFilter)}`;
        }

        const response = await fetch(url);
        const data = await response.json();

        allHistory = (data.history || []).map(item => ({
            ...item.content,
            type: item.type,
            timestamp: item.timestamp,
            session_name: item.session_name || item.content.session_name
        }));

        currentOffset = 0;
        applyFiltersAndSearch();
    } catch (error) {
        console.error('Error loading activity history:', error);
    }
}
```

#### C. Update displayHistory Function

Add notification rendering:

```javascript
if (isNotification) {
    toolLabel = getNotificationIcon(item.notification_type);
    displayText = item.message;

    // Add notification-specific metadata
    if (item.notification_type === 'permission_request' && item.tool_name) {
        metaItems.push(`<div class="command-meta-item">
            <span class="command-meta-label">Tool:</span>
            <span class="command-tool">${escapeHtml(item.tool_name)}</span>
        </div>`);
    }

    if (item.notification_type === 'idle_alert') {
        metaItems.push(`<div class="command-meta-item command-warning">
            ⏱️ Idle
        </div>`);
    }
}

function getNotificationIcon(type) {
    switch(type) {
        case 'permission_request':
            return '🔐 Permission Request';
        case 'idle_alert':
            return '⏱️ Idle Alert';
        default:
            return '🔔 Notification';
    }
}
```

#### D. Add Notification Stats Widget

**New Section**: Notification Insights

```html
<div class="section">
    <h2 class="section-title">Notification Insights</h2>

    <div class="stats-grid">
        <div class="stat-card">
            <div class="stat-label">Permission Requests</div>
            <div class="stat-value" id="permission-requests">0</div>
        </div>

        <div class="stat-card">
            <div class="stat-label">Idle Alerts</div>
            <div class="stat-value" id="idle-alerts">0</div>
        </div>

        <div class="stat-card">
            <div class="stat-label">Most Requested Tool</div>
            <div class="stat-value" id="most-requested-tool">-</div>
            <div class="stat-detail" id="most-requested-count">0 requests</div>
        </div>
    </div>

    <div id="permission-chart"></div>
</div>
```

**Load Stats**:
```javascript
async function loadNotificationStats() {
    try {
        const response = await fetch('/api/notifications/stats');
        const stats = await response.json();

        document.getElementById('permission-requests').textContent = stats.permission_requests;
        document.getElementById('idle-alerts').textContent = stats.idle_alerts;
        document.getElementById('most-requested-tool').textContent = stats.most_requested_tool || 'None';
        document.getElementById('most-requested-count').textContent = `${stats.most_requested_tool_count} requests`;

        // Render chart (e.g., tool permission breakdown)
        renderPermissionChart(stats);
    } catch (error) {
        console.error('Error loading notification stats:', error);
    }
}
```

#### E. Add Permission Request Breakdown Chart

```javascript
async function renderPermissionChart(stats) {
    // Fetch permission breakdown by tool
    const response = await fetch('/api/notifications?type=permission_request&group_by=tool_name');
    const data = await response.json();

    // Simple bar chart or list
    const chartHtml = data.tools.map(tool => `
        <div class="chart-bar">
            <div class="chart-label">${tool.tool_name}</div>
            <div class="chart-bar-bg">
                <div class="chart-bar-fill" style="width: ${tool.percentage}%"></div>
            </div>
            <div class="chart-value">${tool.count}</div>
        </div>
    `).join('');

    document.getElementById('permission-chart').innerHTML = chartHtml;
}
```

#### F. Add CSS Styles

```css
.notification-item {
    background: #fffbf0;
    border-left: 4px solid #f59e0b;
}

.notification-permission {
    background: #fef2f2;
    border-left: 4px solid #ef4444;
}

.notification-idle {
    background: #f0f9ff;
    border-left: 4px solid #3b82f6;
}

.chart-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 8px 0;
}

.chart-label {
    min-width: 100px;
    font-weight: 500;
}

.chart-bar-bg {
    flex: 1;
    height: 24px;
    background: #f3f4f6;
    border-radius: 4px;
    overflow: hidden;
}

.chart-bar-fill {
    height: 100%;
    background: linear-gradient(90deg, #3b82f6, #8b5cf6);
    transition: width 0.3s ease;
}

.chart-value {
    min-width: 40px;
    text-align: right;
    font-weight: 600;
    color: #6b7280;
}
```

#### G. WebSocket Updates

Update WebSocket handler to broadcast notification events:

```javascript
// When notification is received
s.wsHub.Broadcast([]byte(`{"event":"notification_recorded","type":"permission_request"}`))

// Client-side listener
ws.onmessage = function(event) {
    const data = JSON.parse(event.data);

    if (data.event === 'notification_recorded') {
        loadActivityHistory();
        loadNotificationStats();
    }
};
```

### 9. Backend API Implementation Details

**handleRecordNotification**:
```go
func (s *Server) handleRecordNotification(c *fiber.Ctx) error {
    type RecordNotificationRequest struct {
        SessionID        string `json:"session_id"`
        SessionName      string `json:"session_name"`
        NotificationType string `json:"notification_type"`
        Message          string `json:"message"`
        ToolName         string `json:"tool_name"`
        WorkingDirectory string `json:"cwd"`
        GitBranch        string `json:"branch"`
    }

    var req RecordNotificationRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
    }

    notif := &database.Notification{
        ConversationID:   req.SessionID,
        SessionName:      req.SessionName,
        NotificationType: req.NotificationType,
        Message:          req.Message,
        ToolName:         req.ToolName,
        WorkingDirectory: req.WorkingDirectory,
        GitBranch:        req.GitBranch,
        NotifiedAt:       time.Now(),
    }

    if err := s.repo.RecordNotification(notif); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("failed to record notification: %v", err)})
    }

    // Broadcast to WebSocket clients
    s.wsHub.Broadcast([]byte(`{"event":"notification_recorded","type":"` + req.NotificationType + `"}`))

    return c.JSON(fiber.Map{
        "status": "recorded",
        "id":     notif.ID,
        "time":   notif.NotifiedAt,
    })
}
```

**handleGetNotificationStats**:
```go
func (s *Server) handleGetNotificationStats(c *fiber.Ctx) error {
    stats, err := s.repo.GetNotificationStats()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("failed to get stats: %v", err)})
    }

    return c.JSON(stats)
}
```

**Update handleGetAllHistory**:
```go
// Include notifications in unified history
notifications, err := s.repo.GetNotifications(query)
if err != nil {
    return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("failed to get notifications: %v", err)})
}

for _, notif := range notifications {
    allHistory = append(allHistory, HistoryItem{
        Type:             "notification",
        ID:               notif.ID,
        ConversationID:   notif.ConversationID,
        SessionName:      notif.SessionName,
        Timestamp:        notif.NotifiedAt,
        WorkingDirectory: notif.WorkingDirectory,
        GitBranch:        notif.GitBranch,
        Content:          notif,
    })
}
```

### 10. Testing Plan

**Manual Testing**:
1. Install notification hook: `cct --install-notification-hook`
2. Start analytics server: `cct --analytics`
3. Trigger permission requests by using restricted tools
4. Wait 60s to trigger idle alert
5. Verify data appears in dashboard:
   - Notification events in activity history
   - Permission request stats
   - Tool permission breakdown chart
   - Idle alert indicators

**Verify**:
- Notifications appear in unified history timeline
- Stats widget shows correct counts
- Permission chart displays tool breakdown
- Session names match across all activity types
- No blocking or slowdown of Claude Code
- WebSocket updates work in real-time

## Expected Outcome

After implementation:
- Complete activity tracking: prompts + tools + notifications
- Engagement metrics: idle time, response patterns
- Trust evolution: permission request trends over time
- Tool friction analysis: which tools need approval most
- Unified analytics dashboard with all event types
- Session-based filtering across all activities

## Benefits

1. **Engagement Insights** - Track when users are active vs idle
2. **Trust Metrics** - Monitor permission approval patterns
3. **Tool Friction Analysis** - Identify tools that need pre-approval
4. **Productivity Patterns** - Understand response time distributions
5. **Session Quality** - Measure active coding vs waiting time
6. **Permission Recommendations** - Suggest tools to add to allow list

## CLI Usage

```bash
# Install notification hook only
cct --install-notification-hook

# Install all hooks (now includes 3 hooks)
cct --install-all-hooks

# Uninstall notification hook
cct --uninstall-notification-hook

# View analytics with notifications
cct --analytics
```

## Dashboard Features

After implementation, analytics dashboard will show:

### Activity Timeline
- User prompts with session context
- Shell commands with exit codes
- Claude tool usage with parameters
- **Notification events with types** (NEW)

### Notification Insights Widget
- Total permission requests
- Total idle alerts
- Most requested tool
- Permission breakdown chart

### Unified History View
- Filter by: All, Shell, Claude, Prompts, **Notifications** (NEW)
- Session-based filtering
- Search across all event types
- Sortable by timestamp

### Engagement Metrics
- Average response time
- Idle session percentage
- Active coding time
- Permission approval rate

## Technical Notes

### Notification Event Structure (from Claude Code)
```json
{
  "session_id": "uuid",
  "transcript_path": "/path/to/conversation.jsonl",
  "hook_event_name": "Notification",
  "message": "Claude needs your permission to use Bash"
}
```

### Permission Request Message Patterns
```
"Claude needs your permission to use Bash"
"Claude needs your permission to use Read"
"Claude needs your permission to use WebFetch"
```

### Idle Alert Message Pattern
```
"Claude is waiting for your input"
```

### Tool Name Extraction
Use regex to extract tool name from permission messages:
```bash
# From: "Claude needs your permission to use Bash"
# Extract: "Bash"
TOOL_NAME=$(echo "$MESSAGE" | grep -oP '(?<=use )\w+')
```

## Migration Strategy

1. Create notifications table via migration
2. No impact on existing data
3. New hook is optional - won't affect existing hooks
4. Frontend gracefully handles missing notification data
5. Stats show 0 if no notifications recorded yet

## Future Enhancements

1. **Auto-approve suggestions** - "You've approved Bash 50 times, add to allowed?"
2. **Engagement scoring** - Calculate productivity scores
3. **Alert fatigue detection** - Identify frequently blocked tools
4. **Time-of-day patterns** - When are you most responsive?
5. **Session timeout** - Auto-close idle sessions
6. **Permission history timeline** - Track trust evolution over weeks/months
