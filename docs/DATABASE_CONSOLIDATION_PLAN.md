# Database Consolidation Plan

## Executive Summary

This document outlines the plan to consolidate claude-control-terminal's database infrastructure from a fragmented multi-location setup to a **single unified SQLite database** with a cohesive schema and migration system.

---

## Current State Analysis

### Problem: Multiple Database Locations

The application currently creates databases in **TWO different locations**:

```
~/.claude/
├── cct_history.db              ❌ Created by TUI (line 184 in tui/model.go)
└── analytics_data/
    └── cct_history.db          ✅ Created by Server (line 237 in server/server.go)
```

**Root Cause:**
- **TUI**: `dataDir = ~/.claude` (tui/model.go:184)
- **Server**: `dataDir = ~/.claude/analytics_data` (server/server.go:237)
- Both call `database.Initialize(dataDir)` which creates `{dataDir}/cct_history.db`

**Impact:**
- Data fragmentation across two databases
- Provider configs saved in one location might not be visible in another
- Potential data inconsistency

---

### Problem: Split Migration Systems

**Two Separate Migration Systems:**

1. **Main Database Migrations** (`internal/database/database.go:230-413`)
   - Tables: shell_commands, claude_commands, conversations, providers, notifications, etc.
   - 7 migrations currently implemented
   - No version tracking table (uses inline checks)

2. **Agent Table Migrations** (`internal/server/agents/migrations.go:290-328`)
   - Tables: agent_sessions, agent_messages
   - Uses `schema_version` table for tracking
   - Currently at version 4
   - Clean migration system with version control

**Impact:**
- Two separate schema evolution paths
- Risk of version conflicts
- Difficult to coordinate cross-table changes
- Maintenance overhead

---

### Problem: Split Schema Definitions

**Main Schema:** `internal/database/schema.sql` (embedded)
```sql
CREATE TABLE shell_commands (...)
CREATE TABLE claude_commands (...)
CREATE TABLE conversations (...)
CREATE TABLE command_stats (...)
CREATE TABLE user_messages (...)
CREATE TABLE providers (...)
CREATE TABLE notifications (...)
CREATE TABLE user_settings (...)
-- 17 indexes
```

**Agent Schema:** `internal/server/agents/migrations.go` (code-based)
```sql
CREATE TABLE agent_sessions (...)
CREATE TABLE agent_messages (...)
CREATE TABLE schema_version (...)  -- Only for agent tables
-- 6 indexes
```

**Impact:**
- Dual schema maintenance
- No single source of truth
- Harder to visualize complete database structure

---

## Consolidation Goals

1. ✅ **Single Database File**: One `cct_history.db` in one location
2. ✅ **Unified Schema**: All tables defined in one place
3. ✅ **Unified Migrations**: Single migration system with coordinated versioning
4. ✅ **Zero Data Loss**: Migrate existing data from both locations
5. ✅ **Backward Compatibility**: Existing code works with minimal changes

---

## Proposed Solution

### Phase 1: Standardize Database Location

**Decision: Use `~/.claude/analytics_data/cct_history.db` as the canonical location**

**Rationale:**
- Already used by server (primary component)
- Keeps analytics data separate from Claude's own files
- Cleaner directory structure

**Changes Required:**

1. **Fix TUI initialization** (tui/model.go:184)
   ```go
   // Before:
   dataDir := filepath.Join(homeDir, ".claude")

   // After:
   dataDir := filepath.Join(homeDir, ".claude", "analytics_data")
   ```

2. **Create migration helper** to move existing data from old location
   ```go
   func migrateOldDatabaseLocation(oldPath, newPath string) error
   ```

---

### Phase 2: Unified Schema Definition

**Approach: Merge everything into `schema.sql`**

**New Schema Structure:**

```sql
-- ============================================
-- CCT Unified Database Schema
-- Version: 1.0.0
-- ============================================

-- Version tracking (MUST be first)
CREATE TABLE IF NOT EXISTS schema_version (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    version INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial version
INSERT OR IGNORE INTO schema_version (id, version) VALUES (1, 1);

-- ============================================
-- Command History Tables
-- ============================================

CREATE TABLE IF NOT EXISTS shell_commands (...);
CREATE TABLE IF NOT EXISTS claude_commands (...);
CREATE TABLE IF NOT EXISTS command_stats (...);

-- ============================================
-- Conversation & Session Tables
-- ============================================

CREATE TABLE IF NOT EXISTS conversations (...);
CREATE TABLE IF NOT EXISTS user_messages (...);
CREATE TABLE IF NOT EXISTS notifications (...);

-- ============================================
-- Agent Session Tables
-- ============================================

CREATE TABLE IF NOT EXISTS agent_sessions (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'idle',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    message_count INTEGER NOT NULL DEFAULT 0,
    cost_usd REAL NOT NULL DEFAULT 0.0,
    num_turns INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    model_name TEXT,
    claude_session_id TEXT,
    git_branch TEXT,
    options TEXT,  -- JSON-serialized SessionOptions
    CONSTRAINT status_check CHECK (status IN ('idle', 'active', 'processing', 'error', 'ended'))
);

CREATE TABLE IF NOT EXISTS agent_messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    sequence INTEGER NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    thinking_content TEXT,
    tool_uses TEXT,  -- JSON
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tokens_used INTEGER DEFAULT 0,
    FOREIGN KEY (session_id) REFERENCES agent_sessions(id) ON DELETE CASCADE,
    CONSTRAINT role_check CHECK (role IN ('user', 'assistant', 'system'))
);

-- ============================================
-- Configuration Tables
-- ============================================

CREATE TABLE IF NOT EXISTS providers (...);
CREATE TABLE IF NOT EXISTS user_settings (...);

-- ============================================
-- Indexes (All 23+ indexes)
-- ============================================

-- Command indexes
CREATE INDEX IF NOT EXISTS idx_shell_commands_conversation ...;
CREATE INDEX IF NOT EXISTS idx_claude_commands_tool ...;

-- Agent indexes
CREATE INDEX IF NOT EXISTS idx_agent_sessions_status ...;
CREATE INDEX IF NOT EXISTS idx_agent_messages_session ...;

-- ... (all other indexes)
```

---

### Phase 3: Unified Migration System

**New Migration Architecture:**

```
internal/database/
├── database.go           # Main database initialization
├── schema.sql            # Complete unified schema (version 1)
├── migrations.go         # NEW: Unified migration runner
└── migrations/           # NEW: Migration files
    ├── v1_to_v2.sql     # Example: Add new column
    ├── v2_to_v3.sql     # Example: Add new table
    └── ...
```

**Migration Flow:**

```go
// database/migrations.go

type Migration struct {
    Version     int
    Description string
    Up          func(*sql.DB) error
    Down        func(*sql.DB) error  // Optional: rollback
}

var migrations = []Migration{
    {
        Version:     2,
        Description: "Add session_timeout to user_settings",
        Up:          migrateV1ToV2,
    },
    {
        Version:     3,
        Description: "Add cost tracking to conversations",
        Up:          migrateV2ToV3,
    },
    // ... future migrations
}

func runMigrations(db *sql.DB) error {
    currentVersion := getSchemaVersion(db)

    for _, migration := range migrations {
        if migration.Version <= currentVersion {
            continue
        }

        log.Printf("Running migration v%d: %s", migration.Version, migration.Description)

        if err := migration.Up(db); err != nil {
            return fmt.Errorf("migration v%d failed: %w", migration.Version, err)
        }

        if err := setSchemaVersion(db, migration.Version); err != nil {
            return fmt.Errorf("failed to update version to v%d: %w", migration.Version, err)
        }
    }

    return nil
}
```

**Version Tracking:**

- Single `schema_version` table for ALL migrations
- Version numbers are sequential: 1, 2, 3, 4, ...
- Each migration is atomic (wrapped in transaction)
- Failed migrations don't update version number

---

### Phase 4: Data Migration Strategy

**Scenario: User has databases in both locations**

```go
// database/database.go

func Initialize(dataDir string) (*Database, error) {
    // Ensure directory exists
    if err := os.MkdirAll(dataDir, 0755); err != nil {
        return nil, err
    }

    dbPath := filepath.Join(dataDir, "cct_history.db")

    // Check for old database location and migrate if needed
    homeDir, _ := os.UserHomeDir()
    oldDbPath := filepath.Join(homeDir, ".claude", "cct_history.db")

    if _, err := os.Stat(oldDbPath); err == nil && oldDbPath != dbPath {
        // Old database exists at different location
        if _, err := os.Stat(dbPath); err != nil {
            // New location doesn't exist yet - move the file
            log.Printf("Migrating database from %s to %s", oldDbPath, dbPath)
            if err := os.Rename(oldDbPath, dbPath); err != nil {
                // Rename failed, try copy
                if err := copyDatabase(oldDbPath, dbPath); err != nil {
                    return nil, fmt.Errorf("failed to migrate database: %w", err)
                }
                // Keep old DB as backup (don't delete)
                backupPath := oldDbPath + ".backup"
                os.Rename(oldDbPath, backupPath)
            }
        } else {
            // Both locations exist - need to merge
            log.Printf("Found databases in both locations - manual merge required")
            log.Printf("Old: %s", oldDbPath)
            log.Printf("New: %s", dbPath)
            // Use new location and warn user
        }
    }

    // Open database at canonical location
    db, err := sql.Open("sqlite3", dbPath)
    // ... rest of initialization
}
```

---

## Implementation Phases

### Phase 1: Database Location Fix (Week 1)
- [ ] Update TUI to use `analytics_data` directory
- [ ] Add migration helper for old database location
- [ ] Test with existing databases
- [ ] Document migration for users

### Phase 2: Schema Consolidation (Week 1-2)
- [ ] Create unified `schema.sql` with all tables
- [ ] Add comprehensive comments/documentation
- [ ] Move agent table definitions to schema.sql
- [ ] Remove duplicate schema definitions from migrations.go

### Phase 3: Migration System Refactor (Week 2)
- [ ] Create new `migrations.go` with unified runner
- [ ] Port existing main DB migrations to new system
- [ ] Port existing agent migrations to new system
- [ ] Test migration path from v0 (empty) to current
- [ ] Test migration path from existing databases

### Phase 4: Code Updates (Week 2-3)
- [ ] Remove agent-specific migration code
- [ ] Update all database.Initialize() calls
- [ ] Update tests to use unified system
- [ ] Add integration tests for migrations

### Phase 5: Testing & Validation (Week 3)
- [ ] Test with fresh database (v0 → vN)
- [ ] Test with old TUI database (migrate location)
- [ ] Test with old server database (migrate schema)
- [ ] Test with both databases present (merge scenario)
- [ ] Verify all existing functionality works

### Phase 6: Documentation & Release (Week 4)
- [ ] Update README with new database location
- [ ] Create migration guide for existing users
- [ ] Add troubleshooting section
- [ ] Release notes with migration instructions

---

## Migration Commands for Users

**Automatic Migration (Recommended):**
```bash
# Just run CCT - it will automatically migrate
cct --analytics

# Or via TUI
cct
```

**Manual Migration (If Needed):**
```bash
# Backup existing databases
cp ~/.claude/cct_history.db ~/.claude/cct_history.db.backup
cp ~/.claude/analytics_data/cct_history.db ~/.claude/analytics_data/cct_history.db.backup

# Run migration tool (future)
cct migrate --verify
```

---

## Rollback Plan

If issues are discovered after consolidation:

1. **Backup Strategy:**
   - Keep old database files as `.backup`
   - Don't delete during migration
   - Users can manually restore if needed

2. **Version Pinning:**
   - Users can downgrade to previous version
   - Previous version will use old database locations
   - Data won't be lost

3. **Manual Recovery:**
   ```bash
   # Restore old database
   mv ~/.claude/cct_history.db.backup ~/.claude/cct_history.db

   # Use older CCT version
   go install github.com/schlunsen/claude-control-terminal/cmd/cct@v0.x.x
   ```

---

## Breaking Changes

### For Users:
- ❌ **None** - Migration is automatic and transparent

### For Developers:
- ⚠️ Agent migration functions moved from `agents/migrations.go` to `database/migrations.go`
- ⚠️ `InitializeAgentTables()` function removed (integrated into main schema)
- ⚠️ Schema version table format changed (but migrations handle this)

---

## Benefits After Consolidation

1. ✅ **Single Source of Truth**: One database file, one location
2. ✅ **Unified Schema**: Easy to understand complete database structure
3. ✅ **Coordinated Migrations**: No version conflicts between subsystems
4. ✅ **Better Performance**: No duplicate schema checks, single connection pool
5. ✅ **Easier Maintenance**: One migration system to maintain
6. ✅ **Better Testing**: Single database to test against
7. ✅ **Cleaner Codebase**: Less duplication, clearer architecture

---

## Files to Modify

### Core Changes:
1. **`internal/database/schema.sql`** - Add agent tables
2. **`internal/database/database.go`** - Add auto-migration, unified migrations
3. **`internal/database/migrations.go`** - NEW: Unified migration system
4. **`internal/tui/model.go:184`** - Fix dataDir path
5. **`internal/server/agents/migrations.go`** - Remove (migrate logic to database/)
6. **`internal/server/agents/storage.go`** - Update to use unified migrations

### Documentation:
1. **`README.md`** - Update database location
2. **`CHANGELOG.md`** - Document consolidation
3. **`docs/DATABASE_CONSOLIDATION_PLAN.md`** - This document

### Tests:
1. **`internal/database/database_test.go`** - Add migration tests
2. **`internal/database/migrations_test.go`** - NEW: Test all migrations
3. **`internal/server/agents/storage_test.go`** - Update to use unified system

---

## Timeline

- **Total Duration**: 3-4 weeks
- **Critical Path**: Schema consolidation → Migration system → Testing
- **Risk Level**: Medium (data migration always carries risk)
- **Mitigation**: Comprehensive testing, automatic backups, rollback plan

---

## Success Criteria

- [ ] Single database file at `~/.claude/analytics_data/cct_history.db`
- [ ] All 11 tables in unified schema
- [ ] Single migration system with version tracking
- [ ] All existing data preserved after migration
- [ ] All tests pass
- [ ] Zero user-reported data loss
- [ ] Documentation updated

---

## Questions & Decisions

### Q: What happens to `__store.db`?
**A:** Nothing - this is Claude CLI's database, not ours. We don't touch it.

### Q: Should we support multiple database backends (Postgres, MySQL)?
**A:** Not in this phase. SQLite is sufficient for now. Future consideration.

### Q: Should migrations be reversible (Down migrations)?
**A:** Nice to have, but not required for v1. Focus on forward migrations.

### Q: How do we handle schema changes during development?
**A:** Use migration files. Never modify schema.sql directly after v1.

### Q: What if user has massive database (GB scale)?
**A:** SQLite handles this well. If needed, add VACUUM command to maintenance.

---

## Next Steps

1. Review this plan with team
2. Get approval for approach
3. Start with Phase 1 (location fix) - lowest risk
4. Create feature branch: `feature/database-consolidation`
5. Implement phases incrementally
6. Test thoroughly before merging

---

**Last Updated**: 2025-01-XX
**Author**: Claude Code
**Status**: Draft - Awaiting Approval
