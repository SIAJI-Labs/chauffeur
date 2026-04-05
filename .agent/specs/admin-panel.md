# Admin Panel Specification

## Overview

Web-based admin panel served via `chauf serve`, providing a GUI alternative to CLI for managing podman database containers.

## Architecture

### Server (`chauf serve`)

Single Go process serving:
- REST API at `/api/*`
- SSE at `/api/containers/:name/logs`
- Static files (frontend build)

```
chauf serve              # Start in background (default)
chauf serve -f          # Run in foreground
chauf serve --stop      # Stop running server
chauf serve --port 3000 # Custom port (default: 3000)
chauf serve --host panel.test # Custom hostname (default: panel.test)
```

Default: `http://panel.test:3000`

### Daemon Mode

The server runs as a daemon by default:
- PID written to `~/.chauffeur/panel.pid`
- Logs written to `~/.chauffeur/panel.log`
- Detects already running instance
- Graceful shutdown via SIGTERM

### Communication

- **REST API**: JSON over HTTP for CRUD operations
- **Server-Sent Events (SSE)**: One-way server→client for log streaming
- **WebSocket**: Reserved for future bidirectional needs

## API Reference

### Health

```
GET /api/health

Response 200:
{
  "status": "ok",
  "version": "0.1.0",
  "timestamp": "2026-04-05T12:00:00Z"
}
```

### Containers

```
GET /api/containers

Response 200:
[
  {
    "name": "chauf-mysql57",
    "engine": "mysql57",
    "status": "running",
    "hostPort": 3307,
    "createdAt": "2026-01-01T00:00:00Z"
  }
]
```

```
GET /api/containers/:name

Response 200:
{
  "name": "chauf-mysql57",
  "engine": "mysql57",
  "status": "running",
  "hostPort": 3307,
  "createdAt": "2026-01-01T00:00:00Z",
  "config": {
    "DatabasePassword": "******",
    "DatabaseUser": "root",
    "DatabaseName": "app"
  }
}
```

```
POST /api/containers/:name/start

Response 200:
{ "message": "Container started" }

Response 400:
{ "error": "Container already running" }

Response 404:
{ "error": "Container not found" }
```

```
POST /api/containers/:name/stop

Response 200:
{ "message": "Container stopped" }

Response 400:
{ "error": "Container not running" }

Response 404:
{ "error": "Container not found" }
```

```
GET /api/containers/:name/logs

SSE stream of log lines:
data: {"line": 1, "timestamp": "12:00:00", "message": "Starting mysqld..."}
data: {"line": 2, "timestamp": "12:00:01", "message": "Ready for connections"}

Response headers:
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

### Backups

```
GET /api/backups

Response 200:
[
  {
    "name": "chauf-mysql57-2026-04-05-120000.tar.gz",
    "container": "chauf-mysql57",
    "size": 1048576,
    "createdAt": "2026-04-05T12:00:00Z",
    "description": "Manual backup"
  }
]
```

```
POST /api/backups

Body:
{
  "container": "chauf-mysql57",
  "description": "Pre-deployment backup"
}

Response 202:
{ "message": "Backup started", "jobId": "backup-123" }
```

```
POST /api/backups/:name/restore

Body:
{
  "container": "chauf-mysql57",
  " databases": ["app", "test"]
}

Response 202:
{ "message": "Restore started", "jobId": "restore-456" }
```

## WebSocket Protocol

For future bidirectional communication:

```
Client → Server:
{
  "type": "subscribe",
  "target": "container:chauf-mysql57"
}

Server → Client:
{
  "type": "status",
  "target": "container:chauf-mysql57",
  "data": {"status": "running"}
}
```

## Frontend Routes

| Route | Component | Description |
|-------|-----------|-------------|
| `/` | Dashboard | Overview cards: container count, running/stopped, recent backups |
| `/containers` | ContainerList | Table of all containers with start/stop buttons |
| `/containers/:name` | ContainerDetail | Status, config, action buttons |
| `/containers/:name/logs` | ContainerLogs | Live log stream with auto-scroll |
| `/containers/:name/backup` | ContainerBackup | Create restore, select backup |
| `/backups` | BackupList | All backups with size, date, restore action |

## UI Design

### Design System

- **Colors**: shadcn/ui default (slate/zinc palette)
- **Icons**: Lucide React
- **Typography**: System font stack
- **Spacing**: 4px base unit (Tailwind defaults)

### Component Patterns

```tsx
// TanStack Query usage
const { data, isLoading } = useQuery({
  queryKey: ['containers'],
  queryFn: () => api.get('/containers'),
  refetchInterval: 5000, // Poll every 5s for status
})

// Log streaming via SSE
const { addLog, logs } = useLogStream(containerName)
```

### Key UI States

1. **Loading**: Skeleton placeholders matching content shape
2. **Error**: Alert with retry button
3. **Empty**: Illustrated empty state with action hint
4. **Success**: Inline confirmation toast

## File Structure

```
internal/panel/
├── server.go              # Main server, routes, handlers
├── types.go              # Request/response types
├── embed.go              # Static file embedding
└── static/               # Built frontend assets

internal/panel-apps/       # React frontend
├── src/
│   ├── main.tsx
│   ├── App.tsx           # Dashboard with containers list
│   ├── index.css         # Tailwind + theme variables
│   ├── components/ui/
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   └── skeleton.tsx
│   └── lib/
│       ├── utils.ts      # cn() helper
│       └── theme.tsx     # Dark/light theme context
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## Current Implementation

### Implemented Features
- Dashboard with stats cards (total, running, stopped, backups)
- Container list with start/stop buttons
- Loading spinner during actions
- Dark/light theme toggle
- Background daemon mode with PID management

### Pending Features
- Container detail page with logs
- Real-time log streaming via SSE
- Backup/Restore UI
- Backup list page

## Security Considerations

1. **No authentication (v1)**: Local access only, assumes trusted users
2. **CORS**: Disabled by default (same origin)
3. **Rate limiting**: Not implemented in v1
4. **Logging**: All API calls logged to `~/.chauffeur/logs/panel.log`

## Future Enhancements

1. Password protection via `--password` flag
2. Multi-user support with sessions
3. Container terminal (exec into container)
4. Project management (links/unlink)
5. Service management (nginx, PHP-FPM)
