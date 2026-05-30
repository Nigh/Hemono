---
name: hemono-api
description: Hemono (荷物账本) REST API skill for managing shared expense-splitting ledgers, transactions, members, and invitations via /api/v1/ endpoints. Use this skill when the user wants to interact with Hemono programmatically — create ledgers, record expenses, manage members, view statistics, or handle invitation codes.
---

# Hemono API Skill

Hemono (荷物账本) is a shared expense-splitting ledger application. This skill documents the `/api/v1/` REST API, which uses **API token authentication** (not session auth).

## Base URL

```
http://localhost:8090
```

In production, replace with the deployed domain (e.g., `https://yourdomain.com`).

## Authentication

All `/api/v1/*` endpoints require a Bearer token in the `Authorization` header.

```
Authorization: Bearer hmn_<40-hex-chars>
```

Tokens are 44 characters total: `hmn_` prefix + 40 hex characters. They are created via the `/api/tokens` endpoints (which use PocketBase session auth, not covered here).

**If authentication fails**, the API returns:
```json
{ "code": 401, "message": "未认证" }
```

## Amount Convention

All monetary amounts are stored and transmitted as **integer cents (分)**. To display in yuan, divide by 100 and format with 2 decimal places.

- `1000` = ¥10.00
- `1550` = ¥15.50

When creating transactions, always send the amount in cents.

## Quick Reference

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/ledgers` | List user's ledgers |
| `POST` | `/api/v1/ledgers` | Create a new ledger |
| `GET` | `/api/v1/ledgers/{id}` | Get ledger detail |
| `DELETE` | `/api/v1/ledgers/{id}` | Delete ledger (owner only) |
| `GET` | `/api/v1/ledgers/{id}/members` | List ledger members |
| `GET` | `/api/v1/ledgers/{id}/transactions` | List transactions |
| `POST` | `/api/v1/ledgers/{id}/transactions` | Create transaction |
| `DELETE` | `/api/v1/transactions/{id}` | Delete transaction (payer only) |
| `GET` | `/api/v1/ledgers/{id}/stats` | Get ledger statistics |
| `GET` | `/api/v1/ledgers/{id}/invitation` | Get active invitation |
| `POST` | `/api/v1/ledgers/{id}/invitation` | Generate invitation code |
| `DELETE` | `/api/v1/invitations/{id}` | Delete invitation (owner only) |
| `POST` | `/api/v1/invitations/join` | Join ledger via invitation code |

## Common Workflows

### 1. Record an expense split equally (AA)

```bash
# First, list ledgers to get the ledger ID
curl -H "Authorization: Bearer hmn_xxxx" http://localhost:8090/api/v1/ledgers

# Create an AA expense of ¥128.50 (12850 cents) for dinner
curl -X POST http://localhost:8090/api/v1/ledgers/{ledger_id}/transactions \
  -H "Authorization: Bearer hmn_xxxx" \
  -H "Content-Type: application/json" \
  -d '{"amount": 12850, "type": "AA", "direction": "EXPENSE", "note": "晚餐", "date": "2026-05-31"}'
```

### 2. Record a single-beneficiary expense

```bash
# Expense for a specific person (beneficiary = user ID)
curl -X POST http://localhost:8090/api/v1/ledgers/{ledger_id}/transactions \
  -H "Authorization: Bearer hmn_xxxx" \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000, "type": "SINGLE", "direction": "EXPENSE", "beneficiary": "user_id_here", "note": "出租车", "date": "2026-05-31"}'
```

### 3. Invite someone to a ledger

```bash
# Generate invitation code (owner only)
curl -X POST http://localhost:8090/api/v1/ledgers/{ledger_id}/invitation \
  -H "Authorization: Bearer hmn_xxxx" \
  -H "Content-Type: application/json" \
  -d '{"max_uses": 5}'

# Share the code (e.g., ABC-123456) with the other person
# They join using their own token:
curl -X POST http://localhost:8090/api/v1/invitations/join \
  -H "Authorization: Bearer hmn_yyyy" \
  -H "Content-Type: application/json" \
  -d '{"code": "ABC-123456"}'
```

### 4. View ledger statistics

```bash
# All-time stats
curl -H "Authorization: Bearer hmn_xxxx" http://localhost:8090/api/v1/ledgers/{ledger_id}/stats

# Stats for a specific month
curl -H "Authorization: Bearer hmn_xxxx" "http://localhost:8090/api/v1/ledgers/{ledger_id}/stats?month=2026-05"
```

## Error Responses

All errors follow this format:
```json
{
  "code": <http_status_code>,
  "message": "<error_description_in_chinese>"
}
```

| Code | Meaning |
|------|---------|
| 400 | Bad request (invalid parameters) |
| 401 | Unauthorized (missing/invalid token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Resource not found |
| 409 | Conflict (e.g., duplicate invitation, already a member) |
| 410 | Gone (invitation expired) |
| 500 | Internal server error |

## Detailed Documentation

- [API Endpoints](./api-endpoints.md) — Full request/response specs for every endpoint
- [Data Models](./data-models.md) — Collection schemas and field definitions
- [Examples](./examples.md) — Complete curl examples for all operations
