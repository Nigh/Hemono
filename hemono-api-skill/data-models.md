# Data Models

Hemono uses PocketBase collections. Below are the schemas relevant to the `/api/v1/` API.

## Collections

### users

Standard PocketBase auth collection with GitHub OAuth2.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `name` | string | Display name (from GitHub) |
| `email` | string | Email address |
| `avatar` | string | Avatar URL (from GitHub) |

### ledgers

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `name` | string | Ledger name |
| `owner` | string (→ users) | Owner user ID |
| `created` | datetime | Creation timestamp |

**Behavior**: When a ledger is created, the owner is automatically added as an `admin` member in `ledger_members`.

### ledger_members

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `ledger` | string (→ ledgers) | Ledger ID |
| `user` | string (→ users) | User ID |
| `role` | string | Member role (`admin` or `member`) |

**Roles**:
- `admin` — Ledger owner (auto-assigned on creation)
- `member` — Joined via invitation

### transactions

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `ledger` | string (→ ledgers) | Ledger ID |
| `payer` | string (→ users) | User who created the transaction |
| `amount` | int | Amount in cents (分), always > 0 |
| `type` | string | `AA` or `SINGLE` |
| `direction` | string | `EXPENSE` or `INCOME` |
| `beneficiary` | string (→ users) | Required when type = `SINGLE` |
| `note` | string | Transaction description |
| `date` | date | Transaction date (`YYYY-MM-DD`) |
| `created` | datetime | Record creation timestamp |

**Transaction Types**:
- `AA` — Amount is split equally among all ledger members
- `SINGLE` — Amount applies to one specific beneficiary

**Directions**:
- `EXPENSE` — Money spent (payer paid for others)
- `INCOME` — Money received (payer received from others)

### invitation_codes

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `code` | string | Invitation code (`ABC-123456` format) |
| `ledger` | string (→ ledgers) | Target ledger ID |
| `created_by` | string (→ users) | Creator user ID |
| `expires_at` | datetime | Expiration timestamp (24h from creation) |
| `max_uses` | int | Maximum allowed uses (1-99) |
| `used_count` | int | Current usage count |

**Lifecycle**:
- Created with `used_count = 0`
- `used_count` increments on each successful join
- Auto-deleted when `used_count >= max_uses` or accessed after expiry

### api_tokens

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `user` | string (→ users) | Owner user ID |
| `name` | string | Token name (max 100 chars) |
| `token_hash` | string | SHA-256 hash of the token (64 hex chars) |
| `token_prefix` | string | First 12 chars of the token |
| `created` | datetime | Creation timestamp |

**Token format**: `hmn_` + 40 hex chars = 44 chars total. Example: `hmn_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2`

**Storage**: Tokens are stored as SHA-256 hashes. The raw token is only returned once at creation time.

**Limit**: Max 3 tokens per user.

## Entity Relationship

```
users ──< ledgers (owner)
users ──< ledger_members >── ledgers
users ──< transactions (payer)
users ──< transactions (beneficiary, optional)
ledgers ──< transactions
ledgers ──< invitation_codes
users ──< api_tokens
users ──< invitation_codes (created_by)
```

## Balance Semantics

For a given member in a ledger:

```
balance = totalExpense - totalIncome - totalBenefit + incomeShare
```

- **totalExpense**: Sum of all EXPENSE transactions where this user is the payer
- **totalBenefit**: Sum of benefit shares from EXPENSE transactions (AA split or SINGLE targeting this user)
- **totalIncome**: Sum of all INCOME transactions where this user is the payer
- **incomeShare**: Sum of income shares from INCOME transactions (AA split or SINGLE targeting this user)

**Interpretation**:
- `balance > 0` → This member is owed money (paid more than their share)
- `balance < 0` → This member owes money (benefited more than they paid)
- `balance = 0` → Settled
