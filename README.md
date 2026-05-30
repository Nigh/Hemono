# Hemono 荷物账本

## Prepare

```sh
# Create data dir
mkdir pb_data

# Setup PocketBase admin account 
cp .env.example .env
vim .env
```

## Develop

### Docker up
In the dev container, the backend uses Air to automatically monitor changes in Go code, while the frontend relies on Vite's HMR functionality.
```sh
docker compose -f docker-compose.dev.yml up -d
```

### Migrate database

If you change the database configuration, you will need to create a database migration. For details, refer to https://pocketbase.io/docs/go-migrations
```sh
go run . migrate collections
go run . migrate history-sync
```

## Deploy

### Docker up
```sh
docker compose -f docker-compose.yml up -d
```

### Reverse proxy with Caddy
Example Caddyfile

```Caddyfile
yourdomain.com {
    handle /api/* {
        reverse_proxy localhost:8090
    }

    handle /_/* {
        reverse_proxy localhost:8090
    }

    handle {
        reverse_proxy localhost:5173
    }

    encode zstd gzip
}
```

## API Skill (for AI Agents)

This project includes a Kilo skill that enables AI agents to interact with the Hemono `/api/v1/` REST API.

### What's Included

The `hemono-api-skill/` directory contains complete API documentation:

| File | Content |
|------|---------|
| `SKILL.md` | Main skill entry — overview, auth, quick reference, common workflows |
| `api-endpoints.md` | Full request/response specs for every endpoint |
| `data-models.md` | PocketBase collection schemas and field definitions |
| `examples.md` | Complete `curl`, Python, and JavaScript examples |

### How It Works

The skill is auto-discovered by Kilo via `.kilo/skills/hemono-api/SKILL.md`. When a user asks an agent to interact with Hemono (e.g., "帮我记一笔晚餐AA 128元"), the agent loads this skill and uses the documented API to execute the request.

### Prerequisites

1. **API Token**: Create a token via the web UI (Settings → API Tokens) or via `POST /api/tokens` (requires PocketBase session auth). The token format is `hmn_` + 40 hex chars.
2. **Running Server**: The Hemono backend must be running and accessible.

### Using the Skill Manually

You can also reference the API docs directly:

```bash
# List your ledgers
curl -H "Authorization: Bearer hmn_your_token" http://localhost:8090/api/v1/ledgers

# Record a ¥128.50 AA expense
curl -X POST http://localhost:8090/api/v1/ledgers/{id}/transactions \
  -H "Authorization: Bearer hmn_your_token" \
  -H "Content-Type: application/json" \
  -d '{"amount": 12850, "type": "AA", "direction": "EXPENSE", "note": "晚餐"}'
```

### Configuration for Other Projects

To use this skill in another Kilo project, add the skill path to your `kilo.json`:

```jsonc
{
  "skills": {
    "paths": ["./hemono-api-skill"]
  }
}
```

Or copy the `hemono-api-skill/` directory to your project and configure accordingly.
