# AGENTS.md — Hemono (荷物账本)

## Project DNA

**Hemono** is a shared expense-splitting ledger (账本) application. Users create ledgers, record transactions, and settle balances through invitation-based collaboration.

### Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.25 + PocketBase v0.35 (monolith: API + SQLite + Admin UI) |
| **Frontend** | SvelteKit + Svelte 5 (runes) + Tailwind CSS v4 + DaisyUI v5 |
| **Auth** | PocketBase built-in auth + GitHub OAuth2 |
| **Package Mgr** | Bun |
| **Runtime** | Docker Compose (dev: Air hot-reload + Vite HMR; prod: static binaries) |
| **Reverse Proxy** | Caddy (production) |

### Architecture Overview

```
┌──────────────────────────────────────────────┐
│  Frontend (SvelteKit, :5173)                 │
│  ├─ src/routes/          SPA pages           │
│  ├─ src/lib/api/         PocketBase SDK calls│
│  ├─ src/lib/components/  UI components       │
│  ├─ src/lib/stores/      Svelte stores       │
│  └─ src/lib/pb.ts        PB client singleton │
├──────────────────────────────────────────────┤
│  Vite proxy: /api/* , /_/* → backend:8090    │
├──────────────────────────────────────────────┤
│  Backend (PocketBase, :8090)                 │
│  ├─ main.go              Custom routes/hooks │
│  ├─ migrations/          DB schema snapshots │
│  └─ pb_data/             SQLite data dir     │
└──────────────────────────────────────────────┘
```

### Data Model (PocketBase Collections)

- **users** — Auth collection (GitHub OAuth2 mapped: `name`, `avatar`)
- **ledgers** — `name`, `owner` (→ users)
- **ledger_members** — `ledger` (→ ledgers), `user` (→ users), `role`
- **transactions** — `ledger`, `payer` (→ users), `amount` (int, **cents/分**), `type` (`AA` | `SINGLE`), `beneficiary` (→ users), `note`, `date`
- **invitation_codes** — `code` (`ABC-123456`), `ledger`, `created_by`, `expires_at`, `max_uses`, `used_count`

### Custom Backend API Routes (defined in `backend/main.go`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/invitations/by-ledger/{id}` | Get active invitation for a ledger |
| `POST` | `/api/invitations/generate` | Generate new invitation code |
| `GET` | `/api/invitations/by-code/{code}` | Lookup invitation by code |
| `POST` | `/api/invitations/join` | Join ledger via invitation code |
| `DELETE` | `/api/invitations/{id}` | Delete invitation code |
| `GET` | `/api/ledgers/{id}/stats` | Get ledger statistics (supports `?month=YYYY-MM`) |

### Key Design Decisions

- **Amounts stored as integer cents (分)**: `amount = round(yuan * 100)`. All display divides by 100 with `.toFixed(2)`.
- **Transaction types**: `AA` (split equally among all members) vs `SINGLE` (one beneficiary).
- **Invitation codes**: 24h expiry, configurable max uses (1-99), auto-cleanup on access when expired/exhausted.
- **Ledger auto-membership**: `OnRecordAfterCreateSuccess("ledgers")` hook automatically adds creator as `admin` member.
- **Admin bootstrap**: Superuser auto-created from `PB_ADMIN_EMAIL` / `PB_ADMIN_PASSWORD` env vars on bootstrap.

---

## Rules of Engagement

### Frontend

- **Svelte 5 runes only**: Use `$state()`, `$derived()`, `$effect()`, `$props()` — **never** use legacy `let` reactivity or `$:` labels. `GenerateInviteModal.svelte` is a known exception (still uses old syntax); refactor when touching it.
- **Component props**: Define via `interface Props` + `$props()`. Callback props use `on*` naming (e.g., `oncreated`, `onadded`).
- **No `export let`**: This is Svelte 4 syntax. Use `$props()` destructuring instead.
- **Modals**: Use native `<dialog>` element with DaisyUI `modal` class. Open via `dialogElement.showModal()` or `(window as any).id.showModal()`.
- **API calls**: All PocketBase interactions go through `src/lib/api/` modules using the singleton `pb` from `$lib/pb`. Custom routes use `pb.send()`.
- **Notifications**: Use `toasts` store from `$lib/toast` (global). `OperationMessage` component is for inline modal feedback.
- **Styling**: Tailwind CSS v4 + DaisyUI v5. Custom dark theme `xianii` defined in `app.css`. Use DaisyUI semantic classes (`btn`, `card`, `modal`, `alert`, etc.).

### Backend

- **Single-file architecture**: All custom logic lives in `backend/main.go`. Do not split into multiple Go files unless the file exceeds ~1000 lines.
- **PocketBase patterns**: Use `app.FindRecordById`, `FindFirstRecordByFilter`, `FindRecordsByFilter`, `core.NewRecord`, `app.Save`, `app.Delete`.
- **HTTP responses**: Always return structured JSON with `code` + `message` on error. Use `net/http` status constants.
- **Auth check**: Always verify `c.Auth != nil` before processing. Return `401` if unauthenticated.
- **Migrations**: Use PocketBase snapshot migrations. Run `go run . migrate collections` after schema changes, then `go run . migrate history-sync`.

### Code Style

- **Frontend**: Prettier with tabs, single quotes, no trailing commas, 100 char print width. Run `bun run lint` and `bun run check` before considering work done.
- **Backend**: Standard `gofmt` formatting. Keep Chinese comments where they exist for domain context.
- **No comments**: Do not add code comments unless explicitly requested by the user.

### Dev & Deploy

- **Dev**: `docker compose -f docker-compose.dev.yml up -d` — backend uses Air, frontend uses Vite HMR.
- **Prod**: `docker compose -f docker-compose.yml up -d` — builds optimized images.
- **Frontend build**: `bun run build` (SvelteKit adapter-node). Production serves on port 5173 via `bun ./build/index.js`.

---

## Critical Self-Sync Rule

> **When an Agent makes any substantive modification to the project — including but not limited to changes in architecture, core logic, dependency libraries, data model, API routes, or development workflow — the Agent MUST evaluate and update the corresponding sections of this AGENTS.md to ensure it accurately reflects the current state of the project. This rule is non-negotiable.**

Specific triggers for update:
- New or removed PocketBase collections/fields
- New or changed custom API routes
- Dependency version bumps that alter behavior (e.g., Svelte major version, PocketBase breaking changes)
- Changes to dev/deploy scripts or Docker configuration
- New design patterns or coding conventions adopted
- Changes to the auth mechanism or user model
