# User Authentication System

## Overview

This document describes the username/password authentication system added to Claude Control Terminal (CCT). This feature provides secure login functionality for the analytics dashboard and API endpoints.

## Features

### Security
- **Bcrypt Password Hashing**: All passwords are hashed using bcrypt with a cost factor of 12
- **Session-Based Authentication**: Secure session tokens with 24-hour expiration
- **Configurable Access Control**: Choose between open dashboard or login-required mode
- **Admin User Management**: Create, list, and delete users
- **Password Change**: Users can change their passwords

### User Interface
- **Login Modal**: Beautiful, responsive login modal in the frontend
- **TUI Integration**: Setup authentication directly from the TUI
  - Press `U` to configure user authentication
  - Press `O` to open dashboard (automatically prompts for admin setup if needed)
- **Composable Auth**: Vue composable for managing authentication state

## Configuration

### Config File Location
```
~/.claude/analytics/config.json
```

### Configuration Options

```json
{
  "auth": {
    "enabled": true,                    // Legacy API key auth
    "api_key_path": "...",             // API key file path
    "user_auth_enabled": false,        // Enable username/password auth
    "require_login": false,            // Require login for all pages
    "session_timeout_hours": 24        // Session duration (hours)
  }
}
```

### Authentication Modes

1. **Disabled** (default):
   - `user_auth_enabled: false`
   - No authentication required
   - Open access to dashboard and API

2. **Optional Login**:
   - `user_auth_enabled: true`
   - `require_login: false`
   - GET requests allowed without auth
   - POST/PUT/DELETE require authentication
   - Users can browse dashboard without login

3. **Required Login**:
   - `user_auth_enabled: true`
   - `require_login: true`
   - All pages require authentication
   - Login modal shown on page load if not authenticated

## Setup Guide

### Using the TUI

1. **Launch CCT**:
   ```bash
   ./cct
   ```

2. **Enable User Authentication**:
   - Press `U` in the TUI
   - Answer "yes" to enable user authentication
   - Choose whether to require login for all pages
   - Create your admin account (username + password)

3. **Alternative: Automatic Setup**:
   - Press `O` to open the dashboard
   - If user auth is enabled but no admin exists, you'll be prompted to create one

4. **Managing Users**:
   - Press `U` again to:
     - Create additional users
     - List existing users
     - Disable user authentication

### Manual Configuration

1. **Edit Config**:
   ```bash
   vim ~/.claude/analytics/config.json
   ```

2. **Enable User Auth**:
   ```json
   {
     "auth": {
       "user_auth_enabled": true,
       "require_login": true
     }
   }
   ```

3. **Create Admin User** (via API or TUI):
   - Use TUI: Press `U`
   - Or via API (if server running):
     ```bash
     curl -X POST https://localhost:3333/api/auth/users -k \
       -H "Content-Type: application/json" \
       -d '{"username":"admin","password":"your-secure-password","is_admin":true}'
     ```

## API Endpoints

### Public Endpoints (No Auth Required)

```
GET  /api/auth/status      # Check authentication status
POST /api/auth/login       # Login with username/password
GET  /api/health           # Health check
GET  /api/version          # Version info
```

### Authenticated Endpoints

#### User Management (Admin Only)
```
POST   /api/auth/users              # Create new user
GET    /api/auth/users              # List all users
DELETE /api/auth/users/:username    # Delete a user
```

#### Session Management
```
POST /api/auth/logout           # Logout current session
POST /api/auth/change-password  # Change password
```

### API Examples

#### Login
```bash
curl -X POST https://localhost:3333/api/auth/login -k \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'
```

Response:
```json
{
  "token": "session-token-here",
  "username": "admin",
  "expires_at": "2025-01-18T10:00:00Z"
}
```

#### Check Status
```bash
curl https://localhost:3333/api/auth/status -k
```

Response:
```json
{
  "enabled": true,
  "authenticated": true,
  "requireLogin": true,
  "username": "admin",
  "isAdmin": true
}
```

#### Create User (Admin)
```bash
curl -X POST https://localhost:3333/api/auth/users -k \
  -H "Content-Type: application/json" \
  -H "Cookie: session_token=YOUR_SESSION_TOKEN" \
  -d '{"username":"developer","password":"dev-password","is_admin":false}'
```

## File Structure

### Backend Files

```
internal/server/
├── user_auth.go          # User and session management
├── auth_handlers.go      # Authentication HTTP handlers
├── config.go             # Configuration (updated)
└── server.go             # Server setup (updated)

internal/tui/
└── auth_setup.go         # TUI authentication setup
```

### Frontend Files

```
app/
├── components/
│   └── LoginModal.vue         # Login modal component
├── composables/
│   └── useAuth.ts             # Authentication composable
└── layouts/
    └── default.vue            # Layout (updated with login modal)
```

### Data Files

```
~/.claude/analytics/
├── config.json                # Server configuration
├── .secret                    # API key (legacy)
└── users/
    ├── users.json             # User accounts (bcrypt hashed passwords)
    └── sessions.json          # Active sessions
```

## Security Best Practices

### Password Requirements
- Minimum 8 characters
- Stored using bcrypt (cost factor 12)
- Never stored in plain text
- Never logged or exposed in responses

### Session Security
- 256-bit random session tokens
- 24-hour expiration
- HttpOnly cookies (prevents XSS)
- Secure flag (HTTPS only)
- SameSite: Lax

### File Permissions
All authentication files use restricted permissions:
- Config files: `0600` (owner read/write only)
- User files: `0600` (owner read/write only)
- Directory: `0700` (owner access only)

### Best Practices

1. **Use Strong Passwords**:
   - Minimum 12 characters recommended
   - Mix of letters, numbers, and symbols
   - Avoid common passwords

2. **Limit Admin Accounts**:
   - Create one primary admin account
   - Use non-admin accounts for daily use
   - Only create admin accounts when necessary

3. **Regular Password Rotation**:
   - Change passwords periodically
   - Use the "Change Password" feature in settings

4. **Secure the Server**:
   - Keep TLS enabled (`tls.enabled: true`)
   - Bind to localhost only (`host: "127.0.0.1"`)
   - Use SSH tunneling for remote access

5. **Monitor Sessions**:
   - Sessions expire after 24 hours
   - Logout when done using the dashboard
   - Check active sessions regularly

## Migration from API Key Auth

If you're currently using the legacy API key authentication:

1. **Both Can Coexist**:
   - User auth and API key auth are independent
   - Enable user auth without disabling API keys
   - Useful during transition period

2. **Recommended Migration**:
   ```json
   {
     "auth": {
       "enabled": false,           // Disable API key auth
       "user_auth_enabled": true,  // Enable user auth
       "require_login": true       // Require login
     }
   }
   ```

3. **Hooks Still Work**:
   - Analytics hooks continue using API key
   - Frontend uses session authentication
   - No changes needed to hooks

## Troubleshooting

### Cannot Login

1. **Check if user exists**:
   ```bash
   cat ~/.claude/analytics/users/users.json
   ```

2. **Verify authentication is enabled**:
   ```bash
   cat ~/.claude/analytics/config.json | grep user_auth_enabled
   ```

3. **Check server logs** (if verbose mode enabled):
   ```bash
   tail -f ~/.claude/analytics/logs/main.log
   ```

### Forgot Password

Currently, there's no password reset feature. Options:

1. **Delete User and Recreate** (if you have another admin):
   - Use TUI or API to delete user
   - Create new user with same username

2. **Edit Users File** (NOT RECOMMENDED):
   - Stop the server
   - Delete user from `~/.claude/analytics/users/users.json`
   - Restart server and create new user via TUI

3. **Disable User Auth Temporarily**:
   ```json
   {
     "auth": {
       "user_auth_enabled": false
     }
   }
   ```
   - Restart server
   - Access dashboard without auth
   - Re-enable and create new account

### Session Expired

- Sessions last 24 hours by default
- Simply login again to create a new session
- Adjust `session_timeout_hours` in config for longer sessions

### Login Modal Won't Close

- If `require_login: true`, modal cannot be closed
- This is by design - login is required
- Set `require_login: false` for optional login

## Development

### Testing Authentication

1. **Enable User Auth**:
   ```bash
   # Edit config
   vim ~/.claude/analytics/config.json
   ```

2. **Create Test User**:
   ```bash
   ./cct
   # Press U, create user "testuser" with password "testpass123"
   ```

3. **Test Login**:
   - Press A to start server
   - Press O to open browser
   - Login with test credentials

4. **Test API**:
   ```bash
   # Login
   TOKEN=$(curl -X POST https://localhost:3333/api/auth/login -k \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"testpass123"}' \
     | jq -r '.token')

   # Use token
   curl https://localhost:3333/api/stats -k \
     -H "Authorization: Bearer $TOKEN"
   ```

### Building Frontend

```bash
cd internal/server/frontend
npm install
npm run build
```

### Adding New Authenticated Endpoints

```go
// In server.go
auth := api.Group("/auth")
auth.Post("/my-endpoint", s.handleMyEndpoint)

// In handler
func (s *Server) handleMyEndpoint(c *fiber.Ctx) error {
    // Get current user from context
    user, ok := c.Locals("user").(*User)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Not authenticated",
        })
    }

    // Check if admin
    if !user.IsAdmin {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Admin access required",
        })
    }

    // Your logic here
    return c.JSON(fiber.Map{"message": "Success"})
}
```

## Future Enhancements

Potential improvements for future versions:

1. **Password Reset**:
   - Email-based password reset
   - Security questions
   - Admin password override

2. **Multi-Factor Authentication (MFA)**:
   - TOTP support
   - Backup codes

3. **OAuth Integration**:
   - GitHub OAuth
   - Google OAuth
   - Enterprise SSO

4. **Role-Based Access Control (RBAC)**:
   - Custom roles beyond admin/user
   - Fine-grained permissions
   - Team management

5. **Audit Logging**:
   - Login attempts
   - User actions
   - Permission changes

6. **Session Management UI**:
   - View active sessions
   - Revoke sessions remotely
   - Session history

## Support

For issues or questions:

1. Check this documentation first
2. Review server logs with verbose mode enabled
3. Open an issue on GitHub with details
4. Include relevant config (redact passwords!)

## License

MIT License - Same as Claude Control Terminal
