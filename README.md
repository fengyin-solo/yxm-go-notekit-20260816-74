# notekit — Note-Taking Service

A self-contained note-taking service written in pure Go (standard library only).

## Features

- **Notebooks** — group notes under named notebooks
- **Notes** — title, content, tags, pin/archive flags
- **Soft delete** — trash notes with restore support
- **Full-text search** — query notes by title and content
- **Tag filtering** — filter notes by one or more tags
- **Pinned notes** — pin important notes to the top
- **Bearer token authentication**
- **Sliding-window rate limiting per client IP**
- **Atomic JSON file persistence**
- **Full test suite with race detector**

## Architecture

```
cmd/notekit/main.go           — entry point, signal handling, graceful shutdown
internal/httpapi/             — HTTP handlers, routing, JSON responses
internal/service/             — business logic
internal/store/               — in-memory store with atomic JSON persistence
internal/model/               — domain entities (Notebook, Note, NoteFilter)
internal/config/              — environment-based configuration
internal/auth/                — Bearer token authentication
internal/middleware/          — request ID, logging, recovery, rate limiting, timeout
internal/logger/              — leveled structured logger
internal/validator/           — request validation
```

## API Endpoints

### Notebooks

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/notebooks` | List all notebooks |
| POST | `/api/v1/notebooks` | Create a notebook |
| GET | `/api/v1/notebooks/{id}` | Get a notebook |
| PUT | `/api/v1/notebooks/{id}` | Update a notebook |
| DELETE | `/api/v1/notebooks/{id}` | Delete notebook + all its notes |

### Notes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/notes` | List notes (supports filters) |
| POST | `/api/v1/notes` | Create a note |
| GET | `/api/v1/notes/{id}` | Get a note |
| PUT | `/api/v1/notes/{id}` | Update a note |
| DELETE | `/api/v1/notes/{id}` | Permanently delete a note |
| POST | `/api/v1/notes/{id}/trash` | Move to trash (soft delete) |
| POST | `/api/v1/notes/{id}/restore` | Restore from trash |

### Query Parameters for `/api/v1/notes`

| Parameter | Description |
|-----------|-------------|
| `notebook_id=` | Filter by notebook |
| `q=` | Full-text search (title + content) |
| `tag=` | Filter by tag (repeat for multiple) |
| `pinned=true/false` | Filter by pinned status |
| `archived=true/false` | Filter by archived status |
| `include_deleted=true` | Include soft-deleted notes |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `NOTEKIT_ADDR` | `:8080` | Listen address |
| `NOTEKIT_AUTH_TOKEN` | _(empty)_ | Bearer token (empty = disabled) |
| `NOTEKIT_DATA_FILE` | _(empty)_ | Data directory prefix for JSON files |
| `NOTEKIT_SAVE_INTERVAL` | `30s` | Background flush interval |
| `NOTEKIT_RATE_LIMIT` | `100` | Max requests per minute per IP |
| `NOTEKIT_MAX_BODY` | `1048576` | Max request body size |
| `NOTEKIT_TITLE_MAX_BYTES` | `500` | Max title length |

## Build & Run

```bash
# Build
go build ./cmd/notekit

# Run
./notekit
```

## Testing

```bash
go test -race ./...
go vet ./...
```
